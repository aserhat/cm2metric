package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CM2MetricsServer struct {
	server *http.Server
}

func getRepaveMetrics(nodeRepaveMetric *prometheus.GaugeVec) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		configmaps, err := clientset.CoreV1().ConfigMaps("cm2metric").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for _, configmap := range configmaps.Items {
			if strings.HasPrefix(configmap.GetName(), "c2m-node-repave-status") {
				for serverName, repaveStatus := range configmap.Data {
					repaveStatusInt, _ := strconv.ParseFloat(repaveStatus, 64)
					nodeRepaveMetric.With(prometheus.Labels{"hostname": serverName}).Set(repaveStatusInt)
				}
			} else {
				log.Println("no metrics to report")
			}
		}
		time.Sleep(15 * time.Second)
	}

}

func main() {
	nodeRepaveMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "node_repave_status",
			Help: "The repave status of a node, represents the phsae of the repave.",
		},
		[]string{
			"hostname",
		},
	)

	ms := &CM2MetricsServer{
		server: &http.Server{
			Addr: ":8081",
		},
	}
	prometheus.MustRegister(nodeRepaveMetric)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	ms.server.Handler = mux

	go func() {
		go getRepaveMetrics(nodeRepaveMetric)
		if err := ms.server.ListenAndServe(); err != nil {
			log.Println("Filed to listen and serve webhook server: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	prometheus.Unregister(nodeRepaveMetric)
	ms.server.Shutdown(context.Background())
}
