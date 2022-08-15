package slurm

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ExporterErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "error",
			Name:      "slurm_exporter_errors_total",
			Help:      "Total number of Errors from the exporter.",
		},
		[]string{"command", "reason"})
)

func getData(isTest bool, command, file string) string {
	if isTest {
		return readFile(file)
	}
	return execCommand(command)
}

func execCommand(command string) string {
	cmdList := strings.Split(command, " ")
	// TODO: add command context with automatic timeout
	out, err := exec.Command(cmdList[0], cmdList[1:]...).CombinedOutput()
	if err != nil {
		ExporterErrors.WithLabelValues(command, err.Error()).Inc()
		return ""
	}
	return string(out)
}

func readFile(filePath string) string {
	rawData, err := ioutil.ReadFile(filePath)
	if err != nil {
		ExporterErrors.WithLabelValues("readFile", err.Error()).Inc()
		fmt.Println(err)
	}
	return string(rawData)
}

func NewRegistry(gpuCollectorEnabled bool) (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()
	err := reg.Register(NewAccountsCollector()) // from accounts.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewCPUsCollector(false)) // from cpus.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewNodesCollector(false)) // from nodes.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewNodeCollector(false)) // from node.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewPartitionsCollector()) // from partitions.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewQueueCollector(false)) // from queue.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewSchedulerCollector(false)) // from scheduler.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewFairShareCollector()) // from sshare.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewUsersCollector()) // from users.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(ExporterErrors) // from this file
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewScontrolCollector(false)) // from scontrol.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewJobsCollector(false)) // from job.go
	if err != nil {
		return nil, err
	}
	if gpuCollectorEnabled {
		err = reg.Register(NewGPUsCollector()) // from gpus.go
		if err != nil {
			return nil, err
		}
	}
	return reg, nil
}
