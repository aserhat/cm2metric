package main

import (
	"log"
	"net/http"
	"strconv"
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
			log.Println(configmap.GetName(), configmap.Data)
			for serverName, repaveStatus := range configmap.Data {
				repaveStatusInt, _ := strconv.ParseFloat(repaveStatus, 64)
				nodeRepaveMetric.With(prometheus.Labels{"hostname": serverName}).Set(repaveStatusInt)
			}
		}
		time.Sleep(15 * time.Second)
	}

}

func main() {
	nodeRepaveMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "node_repave_alert",
			Help: "The total number iof nodes in a cluster",
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
