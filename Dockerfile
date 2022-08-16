FROM docker-corelib-local.artifactory.mwam.local/ubi8-go-container:0.1.2 as builder

COPY . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o prometheus-slurm-exporter -ldflags="-extldflags=-static"

FROM docker-corelib-local.artifactory.mwam.local/mwam-ubi8:0.1.13
COPY --from=builder /app/prometheus-slurm-exporter /prometheus-slurm-exporter
RUN yum install slurm -y