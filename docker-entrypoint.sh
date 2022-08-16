#!/bin/bash
set -e
# Starting munged which open a socket on /var/run/munge/munge.socket.2 to be used by slurm binaries to authenticate
# The --force just allows to run this with improper permissions on the /var/log/munge directory and log file
/usr/sbin/munged --force

# Starting slurm exporter
/prometheus-slurm-exporter