// Copyright 2017 Victor Penso, Matteo Dessalvi
// SPDX-FileCopyrightText: 2022 2022 Marshall Wace <opensource@mwam.com>
//
// SPDX-License-Identifier: GPL3

package slurm

import (
	"testing"
)

func TestCPUsMetrics(t *testing.T) {
	// Read the input data from a file
	coll := NewCPUsCollector(true)
	t.Logf("%+v", coll.CPUsGetMetrics())
}
