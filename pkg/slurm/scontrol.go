package slurm

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	showNodesDetailsCommand       = "scontrol show -o node"
	showNodesDetailsTestDataInput = "./test_data/scontrol_nodes.txt"
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
			[]string{"name", "arch", "partition", "feature", "address", "version", "os", "state", "weight"}),
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
	}
}

func (s *scontrolCollector) getScontrolMetrics() {
	data := getData(s.isTest, showNodesDetailsCommand, showNodesDetailsTestDataInput)
	lines := strings.Split(data, "\n")
	nodes := []NodeDetails{}
	for _, v := range lines {
		if v == "" {
			continue
		}
		nodes = append(nodes, *parseLine(v))
	}
	for _, n := range nodes {
		for _, partition := range n.Partitions {
			for _, feature := range n.ActiveFeatures {
				s.scontrolNodesInfo.WithLabelValues(n.NodeName, n.Arch, partition, feature, n.NodeAddr, n.Version, n.OS, n.State, n.Weight).Set(1)
			}
		}
		// Now populating all other metrics
		s.scontrolNodeCPUTot.WithLabelValues(n.NodeName).Set(float64(n.CPUTot))
		s.scontrolNodeCPUAllocated.WithLabelValues(n.NodeName).Set(float64(n.CPUAlloc))
		s.scontrolNodeCPULoad.WithLabelValues(n.NodeName).Set(float64(n.CPULoad))
		s.scontrolNodeMemoryTot.WithLabelValues(n.NodeName).Set(float64(n.RealMemory))
		s.scontrolNodeMemoryFree.WithLabelValues(n.NodeName).Set(float64(n.FreeMem))
		s.scontrolNodeMemoryAllocated.WithLabelValues(n.NodeName).Set(float64(n.AllocMem))
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
}

func (s *scontrolCollector) Collect(ch chan<- prometheus.Metric) {
	s.scontrolNodesInfo.Reset()
	s.scontrolNodeCPUAllocated.Reset()
	s.scontrolNodeCPULoad.Reset()
	s.scontrolNodeCPUTot.Reset()
	s.scontrolNodeMemoryAllocated.Reset()
	s.scontrolNodeMemoryFree.Reset()
	s.scontrolNodeMemoryTot.Reset()
	s.getScontrolMetrics()
	s.scontrolNodesInfo.Collect(ch)
	s.scontrolNodeCPUAllocated.Collect(ch)
	s.scontrolNodeCPULoad.Collect(ch)
	s.scontrolNodeCPUTot.Collect(ch)
	s.scontrolNodeMemoryAllocated.Collect(ch)
	s.scontrolNodeMemoryFree.Collect(ch)
	s.scontrolNodeMemoryTot.Collect(ch)
}

func parseLine(node string) *NodeDetails {
	nodedet := &NodeDetails{}

	details := strings.Split(node, " ")
	for _, detail := range details {
		parts := strings.SplitN(detail, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[1] == "" {
			continue
		}
		switch parts[0] {
		case "NodeName":
			nodedet.NodeName = parts[1]
		case "Arch":
			nodedet.Arch = parts[1]
		case "CoresPerSocket":
			nodedet.CoresPerSocket = toUint(parts[1])
		case "CPUAlloc":
			nodedet.CPUAlloc = toUint(parts[1])
		case "CPUTot":
			nodedet.CPUTot = toUint(parts[1])
		case "CPULoad":
			load, _ := strconv.ParseFloat(parts[1], 32)
			nodedet.CPULoad = float32(load)
		case "AvailableFeatures":
			features := toMap(parts[1], ":")
			nodedet.AvailableFeatures = features
		case "ActiveFeatures":
			features := toMap(parts[1], ":")
			nodedet.ActiveFeatures = features
		case "NodeAddr":
			nodedet.NodeAddr = parts[1]
		case "NodeHostName":
			nodedet.NodeHostName = parts[1]
		case "Version":
			nodedet.Version = parts[1]
		case "OS":
			nodedet.OS = parts[1]
		case "RealMemory":
			nodedet.RealMemory = toUint(parts[1])
		case "AllocMem":
			nodedet.AllocMem = toUint(parts[1])
		case "FreeMem":
			nodedet.FreeMem = toUint(parts[1])
		case "Sockets":
			nodedet.Sockets = toUint(parts[1])
		case "Boards":
			nodedet.Boards = toUint(parts[1])
		case "State":
			nodedet.State = parts[1]
		case "ThreadsPerCore":
			nodedet.ThreadsPerCore = toUint(parts[1])
		case "Weight":
			nodedet.Weight = parts[1]
		case "Partitions":
			nodedet.Partitions = strings.Split(parts[1], ",")
		case "CfgTRES":
			tres := toMap(parts[1], "=")
			nodedet.CfgTRES = tres
		case "AllocTRES":
			tres := toMap(parts[1], "=")
			nodedet.AllocTRES = tres
		default:
			continue
		}
	}

	return nodedet
}

func toUint(s string) uint {
	d, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		ExporterErrors.WithLabelValues("convert-string-uint", err.Error()).Inc()
	}
	return uint(d)
}

func toMap(s string, sep string) map[string]string {
	allFeat := strings.Split(s, ",")
	features := map[string]string{}
	for _, feature := range allFeat {
		featurePart := strings.SplitN(feature, sep, 2)
		features[featurePart[0]] = featurePart[1]
	}
	return features
}

type NodeDetails struct {
	NodeName          string
	Arch              string
	CoresPerSocket    uint
	CPUAlloc          uint
	CPUTot            uint
	CPULoad           float32
	AvailableFeatures map[string]string
	ActiveFeatures    map[string]string
	NodeAddr          string
	NodeHostName      string
	Version           string
	OS                string
	RealMemory        uint
	AllocMem          uint
	FreeMem           uint
	Sockets           uint
	Boards            uint
	State             string
	ThreadsPerCore    uint
	Weight            string
	Partitions        []string
	CfgTRES           map[string]string
	AllocTRES         map[string]string
}
