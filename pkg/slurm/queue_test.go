// Copyright 2017 Victor Penso, Matteo Dessalvi

package slurm

import (
	"testing"
)

func TestQueueGetMetrics(t *testing.T) {
	collector := NewQueueCollector(true)
	t.Logf("%+v", collector.QueueGetMetrics())
}
