// Copyright 2017 Victor Penso, Matteo Dessalvi

package slurm

import (
	"testing"
)

func TestCPUsMetrics(t *testing.T) {
	// Read the input data from a file
	coll := NewCPUsCollector(true)
	t.Logf("%+v", coll.CPUsGetMetrics())
}
