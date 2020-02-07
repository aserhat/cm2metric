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
	"k8s.io/client-go/tools/clientcmd"
)

type CM2MetricsServer struct {
	server            *http.Server
	registeredmetrics map[string]*prometheus.GaugeVec
}

func GetKubernetesClient(configType string) *kubernetes.Clientset {
	if configType == "out-of-cluster" {
		config, err := clientcmd.BuildConfigFromFlags("", "/home/godric/.kube/config")
		if err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
		return clientset
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
		return clientset
	}
}

func (c2mserver *CM2MetricsServer) createMetric(metricdetails map[string]string) {
	log.Println("Creating metric: " + metricdetails["prom_metric"])
	nodeRepaveMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricdetails["prom_metric"],
			Help: strings.ReplaceAll(metricdetails["prom_description"], "_", " ") + ".",
		},
		[]string{
			metricdetails["prom_labels"],
		},
	)
	c2mserver.registeredmetrics[metricdetails["prom_metric"]] = nodeRepaveMetric
	log.Println("Registering metric: " + metricdetails["prom_metric"])
	prometheus.MustRegister(nodeRepaveMetric)
}

func (c2mserver *CM2MetricsServer) UpdateRepaveMetrics(clientset *kubernetes.Clientset) {
	for {
		configmaps, err := clientset.CoreV1().ConfigMaps("cm2metric").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, configmap := range configmaps.Items {
			if strings.HasPrefix(configmap.GetName(), "c2m") {
				metricname := configmap.ObjectMeta.Labels["prom_metric"]
				metriclabel := configmap.ObjectMeta.Labels["prom_labels"]
				if metric, ok := c2mserver.registeredmetrics[metricname]; ok {
					log.Println("Recording metric: " + metricname)
					for serverName, repavePhase := range configmap.Data {
						repavePhaseInt, _ := strconv.ParseFloat(repavePhase, 64)
						metric.With(prometheus.Labels{metriclabel: serverName}).Set(repavePhaseInt)
					}
				} else {
					log.Println("Recording metric: " + metricname)
					c2mserver.createMetric(configmap.ObjectMeta.Labels)
					for serverName, repavePhase := range configmap.Data {
						repavePhaseInt, _ := strconv.ParseFloat(repavePhase, 64)
						c2mserver.registeredmetrics[metricname].With(prometheus.Labels{metriclabel: serverName}).Set(repavePhaseInt)
					}
				}
			}
		}
		time.Sleep(15 * time.Second)
	}

}

func main() {
	clientset := GetKubernetesClient("in-cluster")

	registeredmetrics := make(map[string]*prometheus.GaugeVec)

	c2mserver := &CM2MetricsServer{
		server: &http.Server{
			Addr: ":8081",
		},
		registeredmetrics: registeredmetrics,
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	c2mserver.server.Handler = mux

	go func() {
		go c2mserver.UpdateRepaveMetrics(clientset)
		if err := c2mserver.server.ListenAndServe(); err != nil {
			log.Println("Filed to listen and serve c2mserver server: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	//prometheus.Unregister(nodeRepaveMetric)
	c2mserver.server.Shutdown(context.Background())
}
