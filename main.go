/* Copyright 2017-2020 Victor Penso, Matteo Dessalvi, Joeri Hermans

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

package main

import (
	"flag"
	"net/http"

	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vpenso/prometheus-slurm-exporter/pkg/slurm"
)

var listenAddress = flag.String(
	"listen-address",
	":8080",
	"The address to listen on for HTTP requests.")

// Turn on GPUs accounting only if the corresponding command line option is set to true.
var gpuAcct = flag.Bool(
	"gpus-acct",
	false,
	"Enable GPUs accounting")

var execTimeoutSeconds = flag.Int(
	"exec-timeout",
	10,
	"Timeout when executing shell commands")

func main() {
	flag.Parse()

	if *gpuAcct {
		prometheus.MustRegister(slurm.NewGPUsCollector()) // from gpus.go
	}
	reg, err := slurm.NewRegistry(*gpuAcct, *execTimeoutSeconds)
	if err != nil {
		log.Fatalln(err)
	}

	// Adding more collectors to the registry
	reg.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
	log.Printf("Starting Server: %s", *listenAddress)
	log.Printf("GPUs Accounting: %t", *gpuAcct)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
