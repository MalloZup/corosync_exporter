package main

import (
	"flag"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// metrics with labels. (prefer these always as guideline)
	fake = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fake",
			Help: "fake",
		}, []string{"node"})
)

func initMetrics() {
	prometheus.MustRegister(fake)
}

var portNumber = flag.String("port", ":9003", "The port number to listen on for HTTP requests.")
var timeoutSeconds = flag.Int("timeout", 5, "timeout seconds for exporter to wait to fetch new data")

func main() {
	// read cli option and setup initial stat
	flag.Parse()
	initMetrics()
	http.Handle("/metrics", promhttp.Handler())

	// parse each X seconds the cluster configuration and update the metrics accordingly
	// this is done in a goroutine async. we update in this way each 2 second the metrics. (the second will be a parameter in future)
	go func() {
		for {

			// get cluster status xml
			log.Println("[INFO]: Reading corosync information")
			corosyncData, err := exec.Command("corosync-cmapctl").Output()
			if err != nil {
				log.Println("[ERROR]: corosync-cmapctl command execution failed. Did you have corosync-cmapctl installed ?")
				log.Panic(err)
			}
			log.Println("DEBUG: ", corosyncData)
			fake.WithLabelValues("online").Set(float64(1))
			time.Sleep(time.Duration(int64(*timeoutSeconds)) * time.Second)
		}
	}()

	log.Println("[INFO]: Serving metrics on port", *portNumber)
	log.Println("[INFO]: refreshing metric timeouts set to", *timeoutSeconds)
	log.Fatal(http.ListenAndServe(*portNumber, nil))
}
