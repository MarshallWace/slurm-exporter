// Copyright 2020 Victor Penso

package slurm

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type PartitionMetrics struct {
	allocated float64
	idle      float64
	other     float64
	pending   float64
	running   float64
	total     float64
}

func ParsePartitionsMetrics() map[string]*PartitionMetrics {
	partitions := make(map[string]*PartitionMetrics)
	out := execCommand("sinfo -h -o%R,%C")
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, ",") {
			// name of a partition
			partition := strings.Split(line, ",")[0]
			_, key := partitions[partition]
			if !key {
				partitions[partition] = &PartitionMetrics{0, 0, 0, 0, 0, 0}
			}
			states := strings.Split(line, ",")[1]
			allocated, _ := strconv.ParseFloat(strings.Split(states, "/")[0], 64)
			idle, _ := strconv.ParseFloat(strings.Split(states, "/")[1], 64)
			other, _ := strconv.ParseFloat(strings.Split(states, "/")[2], 64)
			total, _ := strconv.ParseFloat(strings.Split(states, "/")[3], 64)
			partitions[partition].allocated = allocated
			partitions[partition].idle = idle
			partitions[partition].other = other
			partitions[partition].total = total
		}
	}
	// get list of pending jobs by partition name
	pendingOut := execCommand("squeue -a -r -h -o%P --states=PENDING")
	list := strings.Split(pendingOut, "\n")
	for _, partition := range list {
		// accumulate the number of pending jobs
		_, key := partitions[partition]
		if key {
			partitions[partition].pending += 1
		}
	}

	// get list of running jobs by partition name
	runningOut := execCommand("squeue -a -r -h -o%P --states=RUNNING")
	list_r := strings.Split(runningOut, "\n")
	for _, partition := range list_r {
		// accumulate the number of running jobs
		_, key := partitions[partition]
		if key {
			partitions[partition].running += 1
		}
	}

	return partitions
}

type PartitionsCollector struct {
	allocated *prometheus.Desc
	idle      *prometheus.Desc
	other     *prometheus.Desc
	pending   *prometheus.Desc
	running   *prometheus.Desc
	total     *prometheus.Desc
}

func NewPartitionsCollector() *PartitionsCollector {
	labels := []string{"partition"}
	return &PartitionsCollector{
		allocated: prometheus.NewDesc("slurm_partition_cpus_allocated", "Allocated CPUs for partition", labels, nil),
		idle:      prometheus.NewDesc("slurm_partition_cpus_idle", "Idle CPUs for partition", labels, nil),
		other:     prometheus.NewDesc("slurm_partition_cpus_other", "Other CPUs for partition", labels, nil),
		pending:   prometheus.NewDesc("slurm_partition_jobs_pending", "Pending jobs for partition", labels, nil),
		running:   prometheus.NewDesc("slurm_partition_jobs_running", "Running jobs for partition", labels, nil),
		total:     prometheus.NewDesc("slurm_partition_cpus_total", "Total CPUs for partition", labels, nil),
	}
}

func (pc *PartitionsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- pc.allocated
	ch <- pc.idle
	ch <- pc.other
	ch <- pc.pending
	ch <- pc.running
	ch <- pc.total
}

func (pc *PartitionsCollector) Collect(ch chan<- prometheus.Metric) {
	pm := ParsePartitionsMetrics()
	for p := range pm {
		if pm[p].allocated > 0 {
			ch <- prometheus.MustNewConstMetric(pc.allocated, prometheus.GaugeValue, pm[p].allocated, p)
		}
		if pm[p].idle > 0 {
			ch <- prometheus.MustNewConstMetric(pc.idle, prometheus.GaugeValue, pm[p].idle, p)
		}
		if pm[p].other > 0 {
			ch <- prometheus.MustNewConstMetric(pc.other, prometheus.GaugeValue, pm[p].other, p)
		}
		if pm[p].pending > 0 {
			ch <- prometheus.MustNewConstMetric(pc.pending, prometheus.GaugeValue, pm[p].pending, p)
		}
		if pm[p].running > 0 {
			ch <- prometheus.MustNewConstMetric(pc.running, prometheus.GaugeValue, pm[p].running, p)
		}
		if pm[p].total > 0 {
			ch <- prometheus.MustNewConstMetric(pc.total, prometheus.GaugeValue, pm[p].total, p)
		}
	}
}
