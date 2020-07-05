# cm2mertics

## Summary
A quick proof of concept to use a ConfigMap to hold metric information and a small app that uses SharedInformers to be notified on Add and Update of configmaps to read the ConfigMap and expose an endpoint serving the metrics.  This could be used in a place for where is not easy to expose a metric but you can make an update to the ConfigMap.  Prometheus can scrape these metrics and make it availble for dashboards and alerting.  One particular place this is used is to track automated server repaving and have the automation pipeline update the ConfigMap along the way.

Current proposed mertic: node_repave_status
Current proposed metric states to follow the build process:

*  0 means server is up and running
*  1 means server is starting a drain.
*  2 means server is starting a rebuild.
*  3 means server is getting prepared after rebuild.
*  4 means server is getting joined to the cluster.
*  5 means server is getting post join configs.

The application registers a Prometheus GaugeVector (GaugeVec) which allows for the repeating of a metric with differnet labels.

## Build and Deploy
Build and publish the images to a container registry.  Might require you to login if the registry you are using is protected.
   Update the version in the VERSION file.
   ```
   vi VERSION (update the number)
   make build
   make contain (make sure to update the registry)
   ```
   Login to your Kubernetes Cluster
   ```
   kubectl apply -f deployment/deployment.yaml (make sure to update the registry
   ```

## Sample ConfigMap
There is a metric registers for node_repave_status and check for ConfigMaps beginning with the name c2m-node-repave-status which drive the update of that metric.  This can be easily adopted to register multiple different metrics and handle differnet ConfigMap naming schemes to handle how to update the metric.
```
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: c2m-node-repave-status-node1
  namespace: cm2metric
data:
  node1: "0"
```
