// Copyright 2017 Victor Penso, Matteo Dessalvi
// SPDX-FileCopyrightText: 2022 2022 Marshall Wace <opensource@mwam.com>
//
// SPDX-License-Identifier: GPL3

package slurm

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	CpuMetricsCommand  = "sinfo -h -o %C"
	CpuMetricsTestData = "test_data/sinfo_cpus.txt"
)

type CPUsMetrics struct {
	alloc float64
	idle  float64
	other float64
	total float64
}

func (cc *CPUsCollector) CPUsGetMetrics() *CPUsMetrics {
	var cm CPUsMetrics
	out := getData(cc.isTest, CpuMetricsCommand, CpuMetricsTestData)
	if strings.Contains(out, "/") {
		splitted := strings.Split(strings.TrimSpace(out), "/")
		cm.alloc, _ = strconv.ParseFloat(splitted[0], 64)
		cm.idle, _ = strconv.ParseFloat(splitted[1], 64)
		cm.other, _ = strconv.ParseFloat(splitted[2], 64)
		cm.total, _ = strconv.ParseFloat(splitted[3], 64)
	}
	return &cm
}

/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm scheduler metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewCPUsCollector(isTest bool) *CPUsCollector {
	return &CPUsCollector{
		isTest: isTest,
		alloc:  prometheus.NewDesc("slurm_cpus_alloc", "Allocated CPUs", nil, nil),
		idle:   prometheus.NewDesc("slurm_cpus_idle", "Idle CPUs", nil, nil),
		other:  prometheus.NewDesc("slurm_cpus_other", "Mix CPUs", nil, nil),
		total:  prometheus.NewDesc("slurm_cpus_total", "Total CPUs", nil, nil),
	}
}

type CPUsCollector struct {
	isTest bool
	alloc  *prometheus.Desc
	idle   *prometheus.Desc
	other  *prometheus.Desc
	total  *prometheus.Desc
}

// Send all metric descriptions
func (cc *CPUsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.alloc
	ch <- cc.idle
	ch <- cc.other
	ch <- cc.total
}
func (cc *CPUsCollector) Collect(ch chan<- prometheus.Metric) {
	cm := cc.CPUsGetMetrics()
	ch <- prometheus.MustNewConstMetric(cc.alloc, prometheus.GaugeValue, cm.alloc)
	ch <- prometheus.MustNewConstMetric(cc.idle, prometheus.GaugeValue, cm.idle)
	ch <- prometheus.MustNewConstMetric(cc.other, prometheus.GaugeValue, cm.other)
	ch <- prometheus.MustNewConstMetric(cc.total, prometheus.GaugeValue, cm.total)
}
