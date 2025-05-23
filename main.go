package main

import (
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/nasshu2916/mirakurun_exporter/collector"
	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"log"
	"net/http"
	"os"
)

var (
	addr                     = kingpin.Flag("addr", "Listen address for web server").Default(":8080").String()
	mirakurunUrl             = kingpin.Flag("mirakurun.url", "Mirakurun URL").Default("http://localhost:40772").String()
	mirakurunRequestTimeout  = kingpin.Flag("mirakurun.request.timeout", "Mirakurun request timeout in seconds").Default("5").Int()
	disableDefaultCollectors = kingpin.Flag("collector.disable-defaults", "Set all collectors to disabled by default.").Default("false").Bool()
)

func main() {
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("node_exporter"))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promslogConfig)

	if *disableDefaultCollectors {
		collector.DisableDefaultCollectors()
	}

	logger.Info("Starting mirakurun_exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	client, err := mirakurun.NewClient(*mirakurunUrl, *mirakurunRequestTimeout)
	if err != nil {
		fmt.Println("Error creating client:", err)
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	http.HandleFunc("/metrics", collector.MetricsHandler(client, logger))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	})

	log.Println("Mirakurun URL:", *mirakurunUrl)
	log.Println("Exporter running on ", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
