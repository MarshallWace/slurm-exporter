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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	showNodesDetailsCommand       = "sinfo -R --json"
	showNodesDetailsTestDataInput = "./test_data/sinfo-nodes.json"
	showNodesDetailsTestDataProm  = "./test_data/sinfo-nodes.prom"
)

var (
	nodeResourcesLabels = []string{"name", "partition"}
)

type nodesCollector struct {
	scontrolNodesInfo           *prometheus.GaugeVec
	scontrolNodeCPULoad         *prometheus.GaugeVec
	scontrolNodeCPUTot          *prometheus.GaugeVec
	scontrolNodeCPUAllocated    *prometheus.GaugeVec
	scontrolNodeMemoryTot       *prometheus.GaugeVec
	scontrolNodeMemoryAllocated *prometheus.GaugeVec
	scontrolNodeMemoryFree      *prometheus.GaugeVec
	scontrolNodeGPUTot          *prometheus.GaugeVec
	scontrolNodeGPUFree         *prometheus.GaugeVec
	isTest                      bool
	nodeAddressSuffix           string

	// Old metrics, keeping them for dashboards/alerts compatibility reasons
	cpuAlloc *prometheus.GaugeVec
	cpuIdle  *prometheus.GaugeVec
	cpuOther *prometheus.GaugeVec
	cpuTotal *prometheus.GaugeVec
	memAlloc *prometheus.GaugeVec
	memTotal *prometheus.GaugeVec
	alloc    *prometheus.GaugeVec
	comp     *prometheus.GaugeVec
	down     *prometheus.GaugeVec
	draining *prometheus.GaugeVec
	drained  *prometheus.GaugeVec
	err      *prometheus.GaugeVec
	fail     *prometheus.GaugeVec
	idle     *prometheus.GaugeVec
	maint    *prometheus.GaugeVec
	mix      *prometheus.GaugeVec
	resv     *prometheus.GaugeVec
}

func NewNodesCollector(isTest bool, nodeAddressSuffix string) *nodesCollector {
	labelsOldmetrics := []string{"node", "status"}
	return &nodesCollector{
		isTest:            isTest,
		nodeAddressSuffix: nodeAddressSuffix,
		scontrolNodesInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_info",
				Help:      "Informations about nodes.",
			},
			[]string{"name", "arch", "partition", "feature", "address", "version", "os", "weight", "state", "reason"}),
		scontrolNodeCPULoad: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_cpu_load",
				Help:      "CPU Load per node as reported by slurm CLI.",
			},
			nodeResourcesLabels),
		scontrolNodeCPUTot: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_cpu_tot",
				Help:      "CPU total available per node as reported by slurm CLI.",
			},
			nodeResourcesLabels),
		scontrolNodeCPUAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_cpu_allocated",
				Help:      "CPU Allocated per node as reported by slurm CLI.",
			},
			nodeResourcesLabels),
		scontrolNodeMemoryTot: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_memory_total_bytes",
				Help:      "Total memory per node as reported by slurm CLI.",
			},
			nodeResourcesLabels),
		scontrolNodeMemoryAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_memory_allocated_bytes",
				Help:      "Allocated memory per node as reported by slurm CLI.",
			},
			nodeResourcesLabels),
		scontrolNodeMemoryFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_memory_free_bytes",
				Help:      "Free memory per node as reported by slurm CLI.",
			},
			nodeResourcesLabels),
		scontrolNodeGPUTot: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_gpu_tot",
				Help:      "Number of total GPU on the node.",
			},
			nodeResourcesLabels),
		scontrolNodeGPUFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_gpu_free",
				Help:      "Number of free GPU on the node.",
			},
			nodeResourcesLabels),

		// Old metrics, keeping them for dashboards/alerts compatibility reasons

		cpuAlloc: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_cpu_alloc",
				Help: "Allocated CPUs per node",
			}, labelsOldmetrics,
		),
		cpuIdle: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_cpu_idle",
				Help: "Idle CPUs per node",
			}, labelsOldmetrics,
		),
		cpuOther: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_cpu_other",
				Help: "Other CPUs per node",
			}, labelsOldmetrics,
		),
		cpuTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_cpu_total",
				Help: "Total CPUs per node",
			}, labelsOldmetrics,
		),
		memAlloc: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_mem_alloc",
				Help: "Allocated memory per node",
			}, labelsOldmetrics,
		),
		memTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_mem_total",
				Help: "Total memory per node",
			}, labelsOldmetrics,
		),
		alloc: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_alloc",
				Help: "Allocated nodes",
			}, []string{},
		),
		comp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_comp",
				Help: "Completing nodes",
			}, []string{},
		),
		down: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_down",
				Help: "Down nodes",
			}, []string{},
		),
		draining: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_draining",
				Help: "Draining nodes",
			}, []string{},
		),
		drained: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_drained",
				Help: "Draining nodes",
			}, []string{},
		),
		err: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_err",
				Help: "Error nodes",
			}, []string{},
		),
		fail: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_fail",
				Help: "Fail nodes",
			}, []string{},
		),
		idle: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_idle",
				Help: "Idle nodes",
			}, []string{},
		),
		maint: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_maint",
				Help: "Maint nodes",
			}, []string{},
		),
		mix: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_mix",
				Help: "Mix nodes",
			}, []string{},
		),
		resv: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_nodes_resv",
				Help: "Reserved nodes",
			}, []string{},
		),
	}
}

func (s *nodesCollector) getNodesMetrics() {
	nodes := &NodeDetails{}
	data := getData(s.isTest, showNodesDetailsCommand, showNodesDetailsTestDataInput)
	err := json.Unmarshal([]byte(data), nodes)
	if err != nil {
		ExporterErrors.WithLabelValues("json-encoding-sinfo-nodes", err.Error()).Inc()
		fmt.Println(err)
	}
	// create metrics from json object
	for _, n := range nodes.Nodes {
		// Prepare state and reason variables
		state := evaluateState(n.State, n.StateFlags)
		reason := ""
		if n.Reason != "" {
			reason = fmt.Sprintf("%s by %s", n.Reason, n.ReasonSetByUser)
		}
		// Iterating over partitions and active_features
		for _, partition := range n.Partitions {
			for _, feature := range strings.Split(n.ActiveFeatures, ",") {
				fqdn := n.Address + s.nodeAddressSuffix
				s.scontrolNodesInfo.WithLabelValues(n.Name, n.Architecture, partition, feature, fqdn, n.SlurmdVersion, n.OperatingSystem, strconv.Itoa(n.Weight), state, reason).Set(1)
			}
			// Now populating all other metrics
			s.scontrolNodeCPUTot.WithLabelValues(n.Name, partition).Set(float64(n.Cpus))
			s.scontrolNodeCPUAllocated.WithLabelValues(n.Name, partition).Set(float64(n.AllocCpus))
			s.scontrolNodeCPULoad.WithLabelValues(n.Name, partition).Set(float64(n.CPULoad) / 100)
			s.scontrolNodeMemoryTot.WithLabelValues(n.Name, partition).Set(float64(n.RealMemory))
			s.scontrolNodeMemoryFree.WithLabelValues(n.Name, partition).Set(float64(n.FreeMemory))
			s.scontrolNodeMemoryAllocated.WithLabelValues(n.Name, partition).Set(float64(n.AllocMemory))
			gpuTot := 0
			if n.Gres != "" {
				// this is a bit obscure, but the standard gres configuration is `gpu:nvidia:3` and we need to get the 3
				gpuTot, err = strconv.Atoi(strings.Split(n.Gres, ":")[2])
				if err != nil {
					ExporterErrors.WithLabelValues("atoi-gpu-tot", err.Error()).Inc()
					fmt.Println(err)
				}

			}
			gpuUsed := 0
			if n.GresUsed != "gpu:0" {
				// this is a bit obscure, but the standard gres configuration is `gpu:nvidia:0(IDX:N\/A)` and we need to get the 0
				gpuUsed, err = strconv.Atoi(strings.Split(strings.Split(n.GresUsed, "(")[0], ":")[2])
				if err != nil {
					ExporterErrors.WithLabelValues("atoi-gpu-used", err.Error()).Inc()
					fmt.Println(err)
				}
			}
			s.scontrolNodeGPUTot.WithLabelValues(n.Name, partition).Set(float64(gpuTot))
			s.scontrolNodeGPUFree.WithLabelValues(n.Name, partition).Set(float64(gpuTot - gpuUsed))

		}

		// Old metrics, keeping them for dashboards/alerts compatibility reasons
		s.cpuAlloc.WithLabelValues(n.Name, state).Set(float64(n.AllocCpus))
		s.cpuIdle.WithLabelValues(n.Name, state).Set(float64(n.IdleCpus))
		s.cpuOther.WithLabelValues(n.Name, state).Set(float64(n.CPUBinding))
		s.cpuTotal.WithLabelValues(n.Name, state).Set(float64(n.Cpus))
		s.memAlloc.WithLabelValues(n.Name, state).Set(float64(n.AllocMemory))
		s.memTotal.WithLabelValues(n.Name, state).Set(float64(n.RealMemory))

		s.aggregateNodeMetrics(state)
	}
}

// aggregateNodeMetrics aggregates metrics https://slurm.schedmd.com/sinfo.html
// This aggregation shoudl be done on prometheus level
// these are deprecated metrics
func (s *nodesCollector) aggregateNodeMetrics(state string) {
	switch state {
	case "ALLOCATED":
		s.alloc.WithLabelValues().Inc()
	case "COMPLETING":
		s.comp.WithLabelValues().Inc()
	case "DOWN":
		s.down.WithLabelValues().Inc()
	case "DRAINING":
		s.draining.WithLabelValues().Inc()
	case "DRAINED":
		s.drained.WithLabelValues().Inc()
	case "FAILING":
		s.err.WithLabelValues().Inc()
	case "FAIL":
		s.fail.WithLabelValues().Inc()
	case "IDLE":
		s.idle.WithLabelValues().Inc()
	case "MAINT":
		s.maint.WithLabelValues().Inc()
	case "MIXED":
		s.mix.WithLabelValues().Inc()
	case "RESERVED":
		s.resv.WithLabelValues().Inc()
	}
}

func evaluateState(state string, stateFlags []string) string {
	if len(stateFlags) == 0 {
		return strings.ToUpper(state)
	}
	stateCombination := strings.ToUpper(state + "-" + strings.Join(stateFlags, "_"))
	switch stateCombination {
	case "IDLE-DRAIN":
		return "DRAINED"
	case "MIXED-DRAIN":
		return "DRAINING"
	case "ALLOCATED-DRAIN":
		return "DRAINING"
	case "DOWN-NOT_RESPONDING":
		return "DOWN"
	default:
		return stateCombination
	}
}

func (s *nodesCollector) Describe(ch chan<- *prometheus.Desc) {
	s.scontrolNodesInfo.Describe(ch)
	s.scontrolNodeCPUAllocated.Describe(ch)
	s.scontrolNodeCPULoad.Describe(ch)
	s.scontrolNodeCPUTot.Describe(ch)
	s.scontrolNodeMemoryAllocated.Describe(ch)
	s.scontrolNodeMemoryFree.Describe(ch)
	s.scontrolNodeMemoryTot.Describe(ch)
	s.scontrolNodeGPUTot.Describe(ch)
	s.scontrolNodeGPUFree.Describe(ch)
	// Old metrics, keeping them for dashboards/alerts compatibility reasons
	s.cpuAlloc.Describe(ch)
	s.cpuIdle.Describe(ch)
	s.cpuOther.Describe(ch)
	s.cpuTotal.Describe(ch)
	s.memAlloc.Describe(ch)
	s.memTotal.Describe(ch)
	s.alloc.Describe(ch)
	s.comp.Describe(ch)
	s.down.Describe(ch)
	s.draining.Describe(ch)
	s.drained.Describe(ch)
	s.err.Describe(ch)
	s.fail.Describe(ch)
	s.idle.Describe(ch)
	s.maint.Describe(ch)
	s.mix.Describe(ch)
	s.resv.Describe(ch)
}

func (s *nodesCollector) Collect(ch chan<- prometheus.Metric) {
	s.scontrolNodesInfo.Reset()
	s.scontrolNodeCPUAllocated.Reset()
	s.scontrolNodeCPULoad.Reset()
	s.scontrolNodeCPUTot.Reset()
	s.scontrolNodeMemoryAllocated.Reset()
	s.scontrolNodeMemoryFree.Reset()
	s.scontrolNodeMemoryTot.Reset()
	s.scontrolNodeGPUFree.Reset()
	s.scontrolNodeGPUTot.Reset()
	// Old metrics, keeping them for dashboards/alerts compatibility reasons
	s.cpuAlloc.Reset()
	s.cpuIdle.Reset()
	s.cpuOther.Reset()
	s.cpuTotal.Reset()
	s.memAlloc.Reset()
	s.memTotal.Reset()
	s.alloc.Reset()
	s.comp.Reset()
	s.down.Reset()
	s.draining.Reset()
	s.drained.Reset()
	s.err.Reset()
	s.fail.Reset()
	s.idle.Reset()
	s.maint.Reset()
	s.mix.Reset()
	s.resv.Reset()
	s.getNodesMetrics()
	s.scontrolNodesInfo.Collect(ch)
	s.scontrolNodeCPUAllocated.Collect(ch)
	s.scontrolNodeCPULoad.Collect(ch)
	s.scontrolNodeCPUTot.Collect(ch)
	s.scontrolNodeMemoryAllocated.Collect(ch)
	s.scontrolNodeMemoryFree.Collect(ch)
	s.scontrolNodeMemoryTot.Collect(ch)
	s.scontrolNodeGPUFree.Collect(ch)
	s.scontrolNodeGPUTot.Collect(ch)
	// Old metrics, keeping them for dashboards/alerts compatibility reasons
	s.cpuAlloc.Collect(ch)
	s.cpuIdle.Collect(ch)
	s.cpuOther.Collect(ch)
	s.cpuTotal.Collect(ch)
	s.memAlloc.Collect(ch)
	s.memTotal.Collect(ch)
	s.alloc.Collect(ch)
	s.comp.Collect(ch)
	s.down.Collect(ch)
	s.draining.Collect(ch)
	s.drained.Collect(ch)
	s.err.Collect(ch)
	s.fail.Collect(ch)
	s.idle.Collect(ch)
	s.maint.Collect(ch)
	s.mix.Collect(ch)
	s.resv.Collect(ch)
}

type NodeDetails struct {
	Meta struct {
		Plugin struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"plugin"`
		Slurm struct {
			Version struct {
				Major int `json:"major"`
				Micro int `json:"micro"`
				Minor int `json:"minor"`
			} `json:"version"`
			Release string `json:"release"`
		} `json:"Slurm"`
	} `json:"meta"`
	Errors []interface{} `json:"errors"`
	Nodes  []struct {
		Architecture              string      `json:"architecture"`
		BurstbufferNetworkAddress string      `json:"burstbuffer_network_address"`
		Boards                    int         `json:"boards"`
		BootTime                  int         `json:"boot_time"`
		Comment                   string      `json:"comment"`
		Cores                     int         `json:"cores"`
		CPUBinding                int         `json:"cpu_binding"`
		CPULoad                   int         `json:"cpu_load"`
		Extra                     string      `json:"extra"`
		FreeMemory                int         `json:"free_memory"`
		Cpus                      int         `json:"cpus"`
		LastBusy                  int         `json:"last_busy"`
		Features                  string      `json:"features"`
		ActiveFeatures            string      `json:"active_features"`
		Gres                      string      `json:"gres"`
		GresDrained               string      `json:"gres_drained"`
		GresUsed                  string      `json:"gres_used"`
		McsLabel                  string      `json:"mcs_label"`
		Name                      string      `json:"name"`
		NextStateAfterReboot      string      `json:"next_state_after_reboot"`
		Address                   string      `json:"address"`
		Hostname                  string      `json:"hostname"`
		State                     string      `json:"state"`
		StateFlags                []string    `json:"state_flags"`
		NextStateAfterRebootFlags []string    `json:"next_state_after_reboot_flags"`
		OperatingSystem           string      `json:"operating_system"`
		Owner                     interface{} `json:"owner"`
		Partitions                []string    `json:"partitions"`
		Port                      int         `json:"port"`
		RealMemory                int         `json:"real_memory"`
		Reason                    string      `json:"reason"`
		ReasonChangedAt           int         `json:"reason_changed_at"`
		ReasonSetByUser           interface{} `json:"reason_set_by_user"`
		SlurmdStartTime           int         `json:"slurmd_start_time"`
		Sockets                   int         `json:"sockets"`
		Threads                   int         `json:"threads"`
		TemporaryDisk             int         `json:"temporary_disk"`
		Weight                    int         `json:"weight"`
		Tres                      string      `json:"tres"`
		SlurmdVersion             string      `json:"slurmd_version"`
		AllocMemory               int         `json:"alloc_memory"`
		AllocCpus                 int         `json:"alloc_cpus"`
		IdleCpus                  int         `json:"idle_cpus"`
		TresUsed                  string      `json:"tres_used"`
		TresWeighted              float64     `json:"tres_weighted"`
	} `json:"nodes"`
}
