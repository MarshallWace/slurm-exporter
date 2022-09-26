// Copyright 2017 Victor Penso, Matteo Dessalvi

package slurm

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	queueCommand  = "squeue -a -r -h -o %A,%T,%r --states=all"
	queueTestData = "./test_data/squeue.txt"
)

type QueueMetrics struct {
	pending       float64
	pending_dep   float64
	running       float64
	suspended     float64
	cancelled     float64
	completing    float64
	completed     float64
	configuring   float64
	failed        float64
	timeout       float64
	preempted     float64
	node_fail     float64
	out_of_memory float64
}

func (qc *QueueCollector) QueueGetMetrics() *QueueMetrics {
	data := getData(qc.isTest, queueCommand, queueTestData)

	var qm QueueMetrics
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.Contains(line, ",") {
			splitted := strings.Split(line, ",")
			state := splitted[1]
			switch state {
			case "PENDING":
				qm.pending++
				if len(splitted) > 2 && splitted[2] == "Dependency" {
					qm.pending_dep++
				}
			case "RUNNING":
				qm.running++
			case "SUSPENDED":
				qm.suspended++
			case "CANCELLED":
				qm.cancelled++
			case "COMPLETING":
				qm.completing++
			case "COMPLETED":
				qm.completed++
			case "CONFIGURING":
				qm.configuring++
			case "FAILED":
				qm.failed++
			case "TIMEOUT":
				qm.timeout++
			case "PREEMPTED":
				qm.preempted++
			case "NODE_FAIL":
				qm.node_fail++
			case "OUT_OF_MEMORY":
				qm.out_of_memory++
			}
		}
	}
	return &qm
}

/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm queue metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewQueueCollector(isTest bool) *QueueCollector {
	return &QueueCollector{
		isTest:        isTest,
		pending:       prometheus.NewDesc("slurm_queue_pending", "Pending jobs in queue", nil, nil),
		pending_dep:   prometheus.NewDesc("slurm_queue_pending_dependency", "Pending jobs because of dependency in queue", nil, nil),
		running:       prometheus.NewDesc("slurm_queue_running", "Running jobs in the cluster", nil, nil),
		suspended:     prometheus.NewDesc("slurm_queue_suspended", "Suspended jobs in the cluster", nil, nil),
		cancelled:     prometheus.NewDesc("slurm_queue_cancelled", "Cancelled jobs in the cluster", nil, nil),
		completing:    prometheus.NewDesc("slurm_queue_completing", "Completing jobs in the cluster", nil, nil),
		completed:     prometheus.NewDesc("slurm_queue_completed", "Completed jobs in the cluster", nil, nil),
		configuring:   prometheus.NewDesc("slurm_queue_configuring", "Configuring jobs in the cluster", nil, nil),
		failed:        prometheus.NewDesc("slurm_queue_failed", "Number of failed jobs", nil, nil),
		timeout:       prometheus.NewDesc("slurm_queue_timeout", "Jobs stopped by timeout", nil, nil),
		preempted:     prometheus.NewDesc("slurm_queue_preempted", "Number of preempted jobs", nil, nil),
		node_fail:     prometheus.NewDesc("slurm_queue_node_fail", "Number of jobs stopped due to node fail", nil, nil),
		out_of_memory: prometheus.NewDesc("slurm_queue_out_of_memory", "Number of jobs stopped by oomkiller", nil, nil),
	}
}

type QueueCollector struct {
	isTest        bool
	pending       *prometheus.Desc
	pending_dep   *prometheus.Desc
	running       *prometheus.Desc
	suspended     *prometheus.Desc
	cancelled     *prometheus.Desc
	completing    *prometheus.Desc
	completed     *prometheus.Desc
	configuring   *prometheus.Desc
	failed        *prometheus.Desc
	timeout       *prometheus.Desc
	preempted     *prometheus.Desc
	node_fail     *prometheus.Desc
	out_of_memory *prometheus.Desc
}

func (qc *QueueCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- qc.pending
	ch <- qc.pending_dep
	ch <- qc.running
	ch <- qc.suspended
	ch <- qc.cancelled
	ch <- qc.completing
	ch <- qc.completed
	ch <- qc.configuring
	ch <- qc.failed
	ch <- qc.timeout
	ch <- qc.preempted
	ch <- qc.node_fail
	ch <- qc.out_of_memory
}

func (qc *QueueCollector) Collect(ch chan<- prometheus.Metric) {
	qm := qc.QueueGetMetrics()
	ch <- prometheus.MustNewConstMetric(qc.pending, prometheus.GaugeValue, qm.pending)
	ch <- prometheus.MustNewConstMetric(qc.pending_dep, prometheus.GaugeValue, qm.pending_dep)
	ch <- prometheus.MustNewConstMetric(qc.running, prometheus.GaugeValue, qm.running)
	ch <- prometheus.MustNewConstMetric(qc.suspended, prometheus.GaugeValue, qm.suspended)
	ch <- prometheus.MustNewConstMetric(qc.cancelled, prometheus.GaugeValue, qm.cancelled)
	ch <- prometheus.MustNewConstMetric(qc.completing, prometheus.GaugeValue, qm.completing)
	ch <- prometheus.MustNewConstMetric(qc.completed, prometheus.GaugeValue, qm.completed)
	ch <- prometheus.MustNewConstMetric(qc.configuring, prometheus.GaugeValue, qm.configuring)
	ch <- prometheus.MustNewConstMetric(qc.failed, prometheus.GaugeValue, qm.failed)
	ch <- prometheus.MustNewConstMetric(qc.timeout, prometheus.GaugeValue, qm.timeout)
	ch <- prometheus.MustNewConstMetric(qc.preempted, prometheus.GaugeValue, qm.preempted)
	ch <- prometheus.MustNewConstMetric(qc.node_fail, prometheus.GaugeValue, qm.node_fail)
	ch <- prometheus.MustNewConstMetric(qc.out_of_memory, prometheus.GaugeValue, qm.out_of_memory)
}
