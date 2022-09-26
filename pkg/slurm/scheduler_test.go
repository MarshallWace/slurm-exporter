// Copyright 2017 Victor Penso, Matteo Dessalvi

package slurm

import (
	"testing"
)

func TestSchedulerGetMetrics(t *testing.T) {
	coll := NewSchedulerCollector(true)
	t.Logf("%+v", coll.SchedulerGetMetrics())
}
