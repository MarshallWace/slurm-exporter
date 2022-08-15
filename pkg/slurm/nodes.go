/* Copyright 2017 Victor Penso, Matteo Dessalvi

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package slurm

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	nodesCommand  = "sinfo -h -o %D,%T"
	nodesTestData = "./test_data/sinfo.txt"
)

type NodesMetrics struct {
	alloc    float64
	comp     float64
	down     float64
	drained  float64
	draining float64
	err      float64
	fail     float64
	idle     float64
	maint    float64
	mix      float64
	resv     float64
}

func RemoveDuplicates(s []string) []string {
	m := map[string]bool{}
	t := []string{}

	// Walk through the slice 's' and for each value we haven't seen so far, append it to 't'.
	for _, v := range s {
		if _, seen := m[v]; !seen {
			if len(v) > 0 {
				t = append(t, v)
				m[v] = true
			}
		}
	}

	return t
}

func (nc *NodesCollector) GetNodesMetrics() *NodesMetrics {
	var nm NodesMetrics
	out := getData(nc.isTest, nodesCommand, nodesTestData)
	lines := strings.Split(out, "\n")

	// Sort and remove all the duplicates from the 'sinfo' output
	sort.Strings(lines)
	lines_uniq := RemoveDuplicates(lines)

	for _, line := range lines_uniq {
		if strings.Contains(line, ",") {
			split := strings.Split(line, ",")
			count, _ := strconv.ParseFloat(strings.TrimSpace(split[0]), 64)
			state := split[1]
			alloc := regexp.MustCompile(`^alloc`)
			comp := regexp.MustCompile(`^comp`)
			down := regexp.MustCompile(`^down`)
			draining := regexp.MustCompile(`^draining`)
			drained := regexp.MustCompile(`^drained`)
			fail := regexp.MustCompile(`^fail`)
			err := regexp.MustCompile(`^err`)
			idle := regexp.MustCompile(`^idle`)
			maint := regexp.MustCompile(`^maint`)
			mix := regexp.MustCompile(`^mix`)
			resv := regexp.MustCompile(`^res`)
			switch {
			case alloc.MatchString(state) == true:
				nm.alloc += count
			case comp.MatchString(state) == true:
				nm.comp += count
			case down.MatchString(state) == true:
				nm.down += count
			case draining.MatchString(state) == true:
				nm.draining += count
			case drained.MatchString(state) == true:
				nm.drained += count
			case fail.MatchString(state) == true:
				nm.fail += count
			case err.MatchString(state) == true:
				nm.err += count
			case idle.MatchString(state) == true:
				nm.idle += count
			case maint.MatchString(state) == true:
				nm.maint += count
			case mix.MatchString(state) == true:
				nm.mix += count
			case resv.MatchString(state) == true:
				nm.resv += count
			}
		}
	}
	return &nm
}

/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm scheduler metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewNodesCollector(isTest bool) *NodesCollector {
	return &NodesCollector{
		isTest:   isTest,
		alloc:    prometheus.NewDesc("slurm_nodes_alloc", "Allocated nodes", nil, nil),
		comp:     prometheus.NewDesc("slurm_nodes_comp", "Completing nodes", nil, nil),
		down:     prometheus.NewDesc("slurm_nodes_down", "Down nodes", nil, nil),
		draining: prometheus.NewDesc("slurm_nodes_draining", "Draining nodes", nil, nil),
		drained:  prometheus.NewDesc("slurm_nodes_drained", "Draining nodes", nil, nil),
		err:      prometheus.NewDesc("slurm_nodes_err", "Error nodes", nil, nil),
		fail:     prometheus.NewDesc("slurm_nodes_fail", "Fail nodes", nil, nil),
		idle:     prometheus.NewDesc("slurm_nodes_idle", "Idle nodes", nil, nil),
		maint:    prometheus.NewDesc("slurm_nodes_maint", "Maint nodes", nil, nil),
		mix:      prometheus.NewDesc("slurm_nodes_mix", "Mix nodes", nil, nil),
		resv:     prometheus.NewDesc("slurm_nodes_resv", "Reserved nodes", nil, nil),
	}
}

type NodesCollector struct {
	isTest   bool
	alloc    *prometheus.Desc
	comp     *prometheus.Desc
	down     *prometheus.Desc
	draining *prometheus.Desc
	drained  *prometheus.Desc
	err      *prometheus.Desc
	fail     *prometheus.Desc
	idle     *prometheus.Desc
	maint    *prometheus.Desc
	mix      *prometheus.Desc
	resv     *prometheus.Desc
}

// Send all metric descriptions
func (nc *NodesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- nc.alloc
	ch <- nc.comp
	ch <- nc.down
	ch <- nc.draining
	ch <- nc.drained
	ch <- nc.err
	ch <- nc.fail
	ch <- nc.idle
	ch <- nc.maint
	ch <- nc.mix
	ch <- nc.resv
}
func (nc *NodesCollector) Collect(ch chan<- prometheus.Metric) {
	nm := nc.GetNodesMetrics()
	ch <- prometheus.MustNewConstMetric(nc.alloc, prometheus.GaugeValue, nm.alloc)
	ch <- prometheus.MustNewConstMetric(nc.comp, prometheus.GaugeValue, nm.comp)
	ch <- prometheus.MustNewConstMetric(nc.down, prometheus.GaugeValue, nm.down)
	ch <- prometheus.MustNewConstMetric(nc.draining, prometheus.GaugeValue, nm.draining)
	ch <- prometheus.MustNewConstMetric(nc.drained, prometheus.GaugeValue, nm.drained)
	ch <- prometheus.MustNewConstMetric(nc.err, prometheus.GaugeValue, nm.err)
	ch <- prometheus.MustNewConstMetric(nc.fail, prometheus.GaugeValue, nm.fail)
	ch <- prometheus.MustNewConstMetric(nc.idle, prometheus.GaugeValue, nm.idle)
	ch <- prometheus.MustNewConstMetric(nc.maint, prometheus.GaugeValue, nm.maint)
	ch <- prometheus.MustNewConstMetric(nc.mix, prometheus.GaugeValue, nm.mix)
	ch <- prometheus.MustNewConstMetric(nc.resv, prometheus.GaugeValue, nm.resv)
}