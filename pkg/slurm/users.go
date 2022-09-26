// Copyright 2020 Victor Penso

package slurm

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type UserJobMetrics struct {
	pending      float64
	running      float64
	running_cpus float64
	suspended    float64
}

func ParseUsersMetrics() map[string]*UserJobMetrics {
	users := make(map[string]*UserJobMetrics)
	out := execCommand("squeue -a -r -h -o %A|%u|%T|%C")
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "|") {
			user := strings.Split(line, "|")[1]
			_, key := users[user]
			if !key {
				users[user] = &UserJobMetrics{0, 0, 0, 0}
			}
			state := strings.Split(line, "|")[2]
			state = strings.ToLower(state)
			cpus, _ := strconv.ParseFloat(strings.Split(line, "|")[3], 64)
			pending := regexp.MustCompile(`^pending`)
			running := regexp.MustCompile(`^running`)
			suspended := regexp.MustCompile(`^suspended`)
			switch {
			case pending.MatchString(state) == true:
				users[user].pending++
			case running.MatchString(state) == true:
				users[user].running++
				users[user].running_cpus += cpus
			case suspended.MatchString(state) == true:
				users[user].suspended++
			}
		}
	}
	return users
}

type UsersCollector struct {
	pending      *prometheus.Desc
	running      *prometheus.Desc
	running_cpus *prometheus.Desc
	suspended    *prometheus.Desc
}

func NewUsersCollector() *UsersCollector {
	labels := []string{"user"}
	return &UsersCollector{
		pending:      prometheus.NewDesc("slurm_user_jobs_pending", "Pending jobs for user", labels, nil),
		running:      prometheus.NewDesc("slurm_user_jobs_running", "Running jobs for user", labels, nil),
		running_cpus: prometheus.NewDesc("slurm_user_cpus_running", "Running cpus for user", labels, nil),
		suspended:    prometheus.NewDesc("slurm_user_jobs_suspended", "Suspended jobs for user", labels, nil),
	}
}

func (uc *UsersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- uc.pending
	ch <- uc.running
	ch <- uc.running_cpus
	ch <- uc.suspended
}

func (uc *UsersCollector) Collect(ch chan<- prometheus.Metric) {
	um := ParseUsersMetrics()
	for u := range um {
		if um[u].pending > 0 {
			ch <- prometheus.MustNewConstMetric(uc.pending, prometheus.GaugeValue, um[u].pending, u)
		}
		if um[u].running > 0 {
			ch <- prometheus.MustNewConstMetric(uc.running, prometheus.GaugeValue, um[u].running, u)
		}
		if um[u].running_cpus > 0 {
			ch <- prometheus.MustNewConstMetric(uc.running_cpus, prometheus.GaugeValue, um[u].running_cpus, u)
		}
		if um[u].suspended > 0 {
			ch <- prometheus.MustNewConstMetric(uc.suspended, prometheus.GaugeValue, um[u].suspended, u)
		}
	}
}
