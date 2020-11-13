// Package metrics handles updating metrics based on ConfigMap values
package metrics

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// MetricsServer exposes a prometheus metrics endpoint
// It uses a SharedInformer to look for ConfigMaps with
// a specific prefix, it reads the labels and updates
// the Prometheus endpoint with a metric representing it.
type MetricsServer struct {
	Server            *http.Server
	Registeredmetrics map[string]*prometheus.GaugeVec
	Clientset         *kubernetes.Clientset
	Informer          cache.SharedIndexInformer
}

const (
	C2M_NAME_PREFIX = "c2m"
)

// Returns a new MetricsServer which holds the HTTP Server to
// expose the metrics endpoint, a map of registered metrics,
// clientset to communicate with the API Server and the Informer
// to listen for ConfigMaps.
func NewServer() *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &MetricsServer{
		Server: &http.Server{
			Addr:    ":8081",
			Handler: mux,
		},
		Registeredmetrics: make(map[string]*prometheus.GaugeVec),
		Clientset:         nil,
		Informer:          nil,
	}
}

// OnAdd checks if a metric is registered based on the ConfigMap
// labels, if not it createds and registers it and then records it
// if it does, it just records it.
func (m *MetricsServer) OnAdd(obj interface{}) {
	configMap := obj.(*corev1.ConfigMap)
	if strings.HasPrefix(configMap.Name, C2M_NAME_PREFIX) {
		metricname := configMap.ObjectMeta.Labels["prom_metric"]
		metriclabel := configMap.ObjectMeta.Labels["prom_labels"]
		if metric, ok := m.Registeredmetrics[metricname]; ok {
			log.Println("Recording metric: " + metricname)
			for serverName, repavePhase := range configMap.Data {
				repavePhaseInt, _ := strconv.ParseFloat(repavePhase, 64)
				metric.With(prometheus.Labels{metriclabel: serverName}).Set(repavePhaseInt)
			}
		} else {
			log.Println("Recording metric: " + metricname)
			m.createMetric(configMap.ObjectMeta.Labels)
			for serverName, repavePhase := range configMap.Data {
				repavePhaseInt, _ := strconv.ParseFloat(repavePhase, 64)
				m.Registeredmetrics[metricname].With(prometheus.Labels{metriclabel: serverName}).Set(repavePhaseInt)
			}
		}
	}
}

// OnUpdate checks if a metric is registered based on the ConfigMap
// labels, if not it createds and registers it and then records it
// if it does, it just records it.
func (m *MetricsServer) OnUpdate(oldObj, obj interface{}) {
	configMap := obj.(*corev1.ConfigMap)
	if strings.HasPrefix(configMap.Name, C2M_NAME_PREFIX) {
		metricname := configMap.ObjectMeta.Labels["prom_metric"]
		metriclabel := configMap.ObjectMeta.Labels["prom_labels"]
		if metric, ok := m.Registeredmetrics[metricname]; ok {
			log.Println("Recording metric: " + metricname)
			for serverName, repavePhase := range configMap.Data {
				repavePhaseInt, _ := strconv.ParseFloat(repavePhase, 64)
				metric.With(prometheus.Labels{metriclabel: serverName}).Set(repavePhaseInt)
			}
		} else {
			log.Println("Recording metric: " + metricname)
			m.createMetric(configMap.ObjectMeta.Labels)
			for serverName, repavePhase := range configMap.Data {
				repavePhaseInt, _ := strconv.ParseFloat(repavePhase, 64)
				m.Registeredmetrics[metricname].With(prometheus.Labels{metriclabel: serverName}).Set(repavePhaseInt)
			}
		}
	}
}

func (m *MetricsServer) createMetric(metricdetails map[string]string) {
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
	m.Registeredmetrics[metricdetails["prom_metric"]] = nodeRepaveMetric
	log.Println("Registering metric: " + metricdetails["prom_metric"])
	prometheus.MustRegister(nodeRepaveMetric)
}
