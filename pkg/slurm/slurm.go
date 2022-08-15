package slurm

import (
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

func execCommand(command string) string {
	cmdList := strings.Split(command, " ")
	out, err := exec.Command(cmdList[0], cmdList[1:]...).CombinedOutput()
	if err != nil {
		ExporterErrors.WithLabelValues(command, err.Error()).Inc()
		return ""
	}
	return string(out)
}
