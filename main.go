package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

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
	prometheus.MustRegister(nodeRepaveMetric)
	go getRepaveMetrics(nodeRepaveMetric)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8081", nil)
}
