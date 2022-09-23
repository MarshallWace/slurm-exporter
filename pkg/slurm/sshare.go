// Copyright 2021 Victor Penso
// SPDX-FileCopyrightText: 2022 2022 Marshall Wace <opensource@mwam.com>
//
// SPDX-License-Identifier: GPL3

package slurm

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type FairShareMetrics struct {
	fairshare float64
}

func ParseFairShareMetrics() map[string]*FairShareMetrics {
	accounts := make(map[string]*FairShareMetrics)
	out := execCommand("sshare -n -P -o account,fairshare")
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "  ") {
			if strings.Contains(line, "|") {
				account := strings.Trim(strings.Split(line, "|")[0], " ")
				_, key := accounts[account]
				if !key {
					accounts[account] = &FairShareMetrics{0}
				}
				fairshare, _ := strconv.ParseFloat(strings.Split(line, "|")[1], 64)
				accounts[account].fairshare = fairshare
			}
		}
	}
	return accounts
}

type FairShareCollector struct {
	fairshare *prometheus.Desc
}

func NewFairShareCollector() *FairShareCollector {
	labels := []string{"account"}
	return &FairShareCollector{
		fairshare: prometheus.NewDesc("slurm_account_fairshare", "FairShare for account", labels, nil),
	}
}

func (fsc *FairShareCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fsc.fairshare
}

func (fsc *FairShareCollector) Collect(ch chan<- prometheus.Metric) {
	fsm := ParseFairShareMetrics()
	for f := range fsm {
		ch <- prometheus.MustNewConstMetric(fsc.fairshare, prometheus.GaugeValue, fsm[f].fairshare, f)
	}
}
