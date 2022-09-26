// Copyright 2017-2022 Victor Penso, Matteo Dessalvi, Joeri Hermans

package main

import (
	"flag"
	"fmt"
	"net/http"

	"log"

	"github.com/MarshallWace/slurm-exporter/pkg/slurm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	appropriateLegalNotice = `
Copyright 2017-2022 Victor Penso, Matteo Dessalvi, Joeri Hermans, Marhall Wace
Victor Penso, Matteo Dessalvi, Joeri Hermans, Marhall Wace LICENSES THE LICENSED SOFTWARE "AS IS," AND MAKES NO EXPRESS OR IMPLIED WARRANTY OF ANY KIND. Victor Penso, Matteo Dessalvi, Joeri Hermans, Marhall Wace SPECIFICALLY DISCLAIMS ALL INDIRECT OR IMPLIED WARRANTIES TO THE FULL EXTENT ALLOWED BY APPLICABLE LAW, INCLUDING WITHOUT LIMITATION ALL IMPLIED WARRANTIES OF, NON-INFRINGEMENT, MERCHANTABILITY, TITLE OR FITNESS FOR ANY PARTICULAR PURPOSE. NO ORAL OR WRITTEN INFORMATION OR ADVICE GIVEN BY Victor Penso, Matteo Dessalvi, Joeri Hermans, Marhall Wace, ITS AGENTS OR EMPLOYEES SHALL CREATE A WARRANTY.
`
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

var nodeAddressSuffix = flag.String(
	"address-suffix",
	"",
	"Suffix to add the node address when reporting metrics")

var ldapServer = flag.String(
	"ldap-address",
	"",
	"Address to contact ldap server. if this is not set, this feature will be disable. if configured please configure --ldap-base-search as well")

var ldapBaseSearch = flag.String(
	"ldap-base-search",
	"",
	"Base search for the ldap server  (e.g. dc=example,dc=com)")

func main() {
	flag.Parse()
	fmt.Println(appropriateLegalNotice)

	if *gpuAcct {
		prometheus.MustRegister(slurm.NewGPUsCollector()) // from gpus.go
	}
	if *ldapServer != "" && *ldapBaseSearch == "" {
		log.Fatalln("--ldap-address is configured but --ldap-base-search is not. please configure --ldap-base-search (e.g. dc=example,dc=com) ")
	}

	reg, err := slurm.NewRegistry(*gpuAcct, *execTimeoutSeconds, *nodeAddressSuffix, *ldapServer, *ldapBaseSearch)
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
