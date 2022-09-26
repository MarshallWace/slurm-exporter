// Copyright 2020 Joeri Hermans, Victor Penso, Matteo Dessalvi

package slurm

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type GPUsMetrics struct {
	alloc       float64
	idle        float64
	total       float64
	utilization float64
}

func GPUsGetMetrics() *GPUsMetrics {
	return ParseGPUsMetrics()
}

func ParseAllocatedGPUs() float64 {
	var num_gpus = 0.0

	output := execCommand("sacct -a -X --format=AllocTRES --state=RUNNING --noheader --parsable2")
	if len(output) > 0 {
		for _, line := range strings.Split(output, "\n") {
			if len(line) > 0 {
				line = strings.Trim(line, "\"")
				for _, resource := range strings.Split(line, ",") {
					if strings.HasPrefix(resource, "gres/gpu=") {
						descriptor := strings.TrimPrefix(resource, "gres/gpu=")
						job_gpus, _ := strconv.ParseFloat(descriptor, 64)
						num_gpus += job_gpus
					}
				}
			}
		}
	}

	return num_gpus
}

func ParseTotalGPUs() float64 {
	var num_gpus = 0.0
	out := execCommand("sinfo -h -o \"%n %G\"")
	if len(out) > 0 {
		for _, line := range strings.Split(out, "\n") {
			if len(line) > 0 {
				line = strings.Trim(line, "\"")
				gres := strings.Fields(line)[1]
				// gres column format: comma-delimited list of resources
				for _, resource := range strings.Split(gres, ",") {
					if strings.HasPrefix(resource, "gpu:") {
						// format: gpu:<type>:N(S:<something>), e.g. gpu:RTX2070:2(S:0)
						descriptor := strings.Split(resource, ":")[2]
						descriptor = strings.Split(descriptor, "(")[0]
						node_gpus, _ := strconv.ParseFloat(descriptor, 64)
						num_gpus += node_gpus
					}
				}
			}
		}
	}

	return num_gpus
}

func ParseGPUsMetrics() *GPUsMetrics {
	var gm GPUsMetrics
	total_gpus := ParseTotalGPUs()
	allocated_gpus := ParseAllocatedGPUs()
	gm.alloc = allocated_gpus
	gm.idle = total_gpus - allocated_gpus
	gm.total = total_gpus
	gm.utilization = allocated_gpus / total_gpus
	return &gm
}

/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm scheduler metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewGPUsCollector() *GPUsCollector {
	return &GPUsCollector{
		alloc:       prometheus.NewDesc("slurm_gpus_alloc", "Allocated GPUs", nil, nil),
		idle:        prometheus.NewDesc("slurm_gpus_idle", "Idle GPUs", nil, nil),
		total:       prometheus.NewDesc("slurm_gpus_total", "Total GPUs", nil, nil),
		utilization: prometheus.NewDesc("slurm_gpus_utilization", "Total GPU utilization", nil, nil),
	}
}

type GPUsCollector struct {
	alloc       *prometheus.Desc
	idle        *prometheus.Desc
	total       *prometheus.Desc
	utilization *prometheus.Desc
}

// Send all metric descriptions
func (cc *GPUsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.alloc
	ch <- cc.idle
	ch <- cc.total
	ch <- cc.utilization
}
func (cc *GPUsCollector) Collect(ch chan<- prometheus.Metric) {
	cm := GPUsGetMetrics()
	ch <- prometheus.MustNewConstMetric(cc.alloc, prometheus.GaugeValue, cm.alloc)
	ch <- prometheus.MustNewConstMetric(cc.idle, prometheus.GaugeValue, cm.idle)
	ch <- prometheus.MustNewConstMetric(cc.total, prometheus.GaugeValue, cm.total)
	ch <- prometheus.MustNewConstMetric(cc.utilization, prometheus.GaugeValue, cm.utilization)
}
