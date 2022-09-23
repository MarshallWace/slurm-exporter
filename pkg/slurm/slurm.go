/* Copyright 2022 Marshall Wace Asset Management

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
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"github.com/MarshallWace/slurm-exporter/pkg/ldapsearch"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ExporterErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "",
			Name:      "slurm_exporter_errors_total",
			Help:      "Total number of Errors from the exporter.",
		},
		[]string{"command", "reason"})
	ExecDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "",
			Name:      "slurm_exporter_exec_duration",
			Help:      "Duration of exec commands.",
		},
		[]string{"command"})
	execTimeoutSeconds = 10
)

func getData(isTest bool, command, file string) string {
	if isTest {
		return readFile(file)
	}
	return execCommand(command)
}

func execCommand(command string) string {
	cmdList := strings.Split(command, " ")
	before := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(execTimeoutSeconds)*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, cmdList[0], cmdList[1:]...).Output()
	elapsed := time.Since(before)
	ExecDuration.WithLabelValues(command).Observe(elapsed.Seconds())
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

func NewRegistry(gpuCollectorEnabled bool, exectimeout int, nodeAddressSuffix, ldapServer, ldapBaseSearch string) (*prometheus.Registry, error) {
	reg := prometheus.NewRegistry()
	execTimeoutSeconds = exectimeout
	err := reg.Register(NewAccountsCollector()) // from accounts.go
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewCPUsCollector(false)) // from cpus.go
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
	err = reg.Register(ExecDuration) // from this file
	if err != nil {
		return nil, err
	}
	err = reg.Register(NewNodesCollector(false, nodeAddressSuffix)) // from scontrol.go
	if err != nil {
		return nil, err
	}
	if ldapServer != "" {
		ldap, err := ldapsearch.Init("", ldapServer, ldapBaseSearch)
		if err != nil {
			ExporterErrors.WithLabelValues("ldapsearch", err.Error()).Inc()
			fmt.Println(err)
		}
		err = reg.Register(NewJobsCollector(false, ldap)) // from jobs.go
		if err != nil {
			return nil, err
		}
	} else {
		err = reg.Register(NewJobsCollector(false, nil)) // from jobs.go
		if err != nil {
			return nil, err
		}
	}
	if gpuCollectorEnabled {
		err = reg.Register(NewGPUsCollector()) // from gpus.go
		if err != nil {
			return nil, err
		}
	}

	return reg, nil
}
