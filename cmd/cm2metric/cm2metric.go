package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/aserhat/cm2metric/internal/pkg/metrics"
)

func main() {
	log.Print("Configmap To Metrics (c2m) server has started.")

	// Get our clientset
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		log.Panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		log.Panic(err.Error())
	}

	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().ConfigMaps().Informer()

	// Get a new metrics server and set the clientset and informer
	c2mserver := metrics.NewServer()
	c2mserver.Clientset = clientset
	c2mserver.Informer = informer

	// Setup the Event Handler for Informer
	c2mserver.Informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c2mserver.OnAdd,
		UpdateFunc: c2mserver.OnUpdate,
		DeleteFunc: c2mserver.OnDelete,
	})

	// Start and run the metrics server
	go func() {
		if err := c2mserver.Server.ListenAndServe(); err != nil {
			log.Printf("Filed to listen and serve c2mserver server: %v", err)
		}
	}()

	// Handle stopping the server
	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()

	// Start the informer
	go c2mserver.Informer.Run(stopper)

	if !cache.WaitForCacheSync(stopper, c2mserver.Informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	<-stopper

	if err2 := c2mserver.Server.Shutdown(context.Background()); err2 != nil {
		log.Printf("Filed to listen and serve c2mserver server: %v", err2)
	}
}
