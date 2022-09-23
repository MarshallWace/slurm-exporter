// Copyright 2017 Victor Penso, Matteo Dessalvi
// SPDX-FileCopyrightText: 2022 2022 Marshall Wace <opensource@mwam.com>
//
// SPDX-License-Identifier: GPL3

package slurm

import (
	"testing"
)

func TestQueueGetMetrics(t *testing.T) {
	collector := NewQueueCollector(true)
	t.Logf("%+v", collector.QueueGetMetrics())
}
