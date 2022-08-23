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
	showNodesDetailsTestDataInput = "./test_data/sinfo-reason-nodes.json"
	showNodesDetailsTestDataProm  = "./test_data/scontrol_nodes.prom"
)

type scontrolCollector struct {
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
}

func NewScontrolCollector(isTest bool) *scontrolCollector {
	return &scontrolCollector{
		isTest: isTest,
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
			[]string{"name"}),
		scontrolNodeCPUTot: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_cpu_tot",
				Help:      "CPU total available per node as reported by slurm CLI.",
			},
			[]string{"name"}),
		scontrolNodeCPUAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_cpu_allocated",
				Help:      "CPU Allocated per node as reported by slurm CLI.",
			},
			[]string{"name"}),
		scontrolNodeMemoryTot: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_memory_total_bytes",
				Help:      "Total memory per node as reported by slurm CLI.",
			},
			[]string{"name"}),
		scontrolNodeMemoryAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_memory_allocated_bytes",
				Help:      "Allocated memory per node as reported by slurm CLI.",
			},
			[]string{"name"}),
		scontrolNodeMemoryFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_memory_free_bytes",
				Help:      "Free memory per node as reported by slurm CLI.",
			},
			[]string{"name"}),
		scontrolNodeGPUTot: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_gpu_tot",
				Help:      "Number of total GPU on the node.",
			},
			[]string{"name"}),
		scontrolNodeGPUFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_node_gpu_free",
				Help:      "Number of free GPU on the node.",
			},
			[]string{"name"}),
	}
}

func (s *scontrolCollector) getScontrolMetrics() {
	nodes := &NodeDetails{}
	data := getData(s.isTest, showNodesDetailsCommand, showNodesDetailsTestDataInput)
	err := json.Unmarshal([]byte(data), nodes)
	if err != nil {
		ExporterErrors.WithLabelValues("json-encoding-sinfo-nodes", err.Error()).Inc()
		fmt.Println(err)
	}
	// create metrics from json object
	for _, n := range nodes.Nodes {
		for _, partition := range n.Partitions {
			for _, feature := range strings.Split(n.ActiveFeatures, ",") {
				state := n.State
				if len(n.StateFlags) > 0 {
					state = strings.Join(n.StateFlags, "-")
				}
				reason := ""
				if n.Reason != "" {
					reason = fmt.Sprintf("%s by %s", n.Reason, n.ReasonSetByUser)
				}
				s.scontrolNodesInfo.WithLabelValues(n.Name, n.Architecture, partition, feature, n.Address, n.SlurmdVersion, n.OperatingSystem, strconv.Itoa(n.Weight), state, reason).Set(1)
			}
		}
		// Now populating all other metrics
		s.scontrolNodeCPUTot.WithLabelValues(n.Name).Set(float64(n.Cpus))
		s.scontrolNodeCPUAllocated.WithLabelValues(n.Name).Set(float64(n.AllocCpus))
		s.scontrolNodeCPULoad.WithLabelValues(n.Name).Set(float64(n.CPULoad) / 100)
		s.scontrolNodeMemoryTot.WithLabelValues(n.Name).Set(float64(n.RealMemory))
		s.scontrolNodeMemoryFree.WithLabelValues(n.Name).Set(float64(n.FreeMemory))
		s.scontrolNodeMemoryAllocated.WithLabelValues(n.Name).Set(float64(n.AllocMemory))
		gpuTot := 0
		if n.Gres != "" {
			// this is a bit obscure, but the standard gres configuration is `gpu:nvidia:3` and we need to get the 3
			gpuTot, _ = strconv.Atoi(strings.Split(n.Gres, ":")[2])
		}
		gpuUsed := 0
		if n.GresUsed != "gpu:0" {
			// this is a bit obscure, but the standard gres configuration is `gpu:nvidia:0(IDX:N\/A)` and we need to get the 0
			gpuUsed, _ = strconv.Atoi(strings.Split(strings.Split(n.GresUsed, "(")[0], ":")[2])
		}
		s.scontrolNodeGPUTot.WithLabelValues(n.Name).Set(float64(gpuTot))
		s.scontrolNodeGPUFree.WithLabelValues(n.Name).Set(float64(gpuTot - gpuUsed))
	}
}

func (s *scontrolCollector) Describe(ch chan<- *prometheus.Desc) {
	s.scontrolNodesInfo.Describe(ch)
	s.scontrolNodeCPUAllocated.Describe(ch)
	s.scontrolNodeCPULoad.Describe(ch)
	s.scontrolNodeCPUTot.Describe(ch)
	s.scontrolNodeMemoryAllocated.Describe(ch)
	s.scontrolNodeMemoryFree.Describe(ch)
	s.scontrolNodeMemoryTot.Describe(ch)
	s.scontrolNodeGPUTot.Describe(ch)
	s.scontrolNodeGPUFree.Describe(ch)
}

func (s *scontrolCollector) Collect(ch chan<- prometheus.Metric) {
	s.scontrolNodesInfo.Reset()
	s.scontrolNodeCPUAllocated.Reset()
	s.scontrolNodeCPULoad.Reset()
	s.scontrolNodeCPUTot.Reset()
	s.scontrolNodeMemoryAllocated.Reset()
	s.scontrolNodeMemoryFree.Reset()
	s.scontrolNodeMemoryTot.Reset()
	s.scontrolNodeGPUFree.Reset()
	s.scontrolNodeGPUTot.Reset()
	s.getScontrolMetrics()
	s.scontrolNodesInfo.Collect(ch)
	s.scontrolNodeCPUAllocated.Collect(ch)
	s.scontrolNodeCPULoad.Collect(ch)
	s.scontrolNodeCPUTot.Collect(ch)
	s.scontrolNodeMemoryAllocated.Collect(ch)
	s.scontrolNodeMemoryFree.Collect(ch)
	s.scontrolNodeMemoryTot.Collect(ch)
	s.scontrolNodeGPUFree.Collect(ch)
	s.scontrolNodeGPUTot.Collect(ch)
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
