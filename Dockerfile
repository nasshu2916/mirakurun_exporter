ARG GO_VERSION=1.24.1

FROM golang:${GO_VERSION} AS build
COPY . .
RUN CGO_ENABLED=0 go build -o /mirakurun_exporter

FROM alpine:latest
COPY --from=build /mirakurun_exporter /bin/mirakurun_exporter
ENTRYPOINT ["/bin/mirakurun_exporter"]
CMD []
