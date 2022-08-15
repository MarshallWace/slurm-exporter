FROM docker-corelib-local.artifactory.mwam.local/ubi8-go-container:0.1.2 as builder

COPY . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o prometheus-slurm-exporter -ldflags="-extldflags=-static"

FROM docker-dockerio-remote.artifactory.mwam.local/library/alpine:3.16.2
COPY --from=builder /app/prometheus-slurm-exporter /prometheus-slurm-exporter