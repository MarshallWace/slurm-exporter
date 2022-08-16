#!/env/bash

pipefail -x

# Starting munged which open a socket on /var/run/munge/munge.socket.2 to be used by slurm binaries to authenticate
/usr/sbin/munged

# Starting slurm exporter
/prometheus-slurm-exporter