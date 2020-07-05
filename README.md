# cm2mertics

## Summary
A quick proof of concept which allows you to use ConfigMaps to hold metric information.  This app uses SharedInformers to be notified on Add and Update of ConfigMaps to read the ConfigMap and expose the data in it as a metric via a /metrics endpoint.  This could be used in a place for where you want metrics around a process such the stages of a pipeline.  

## Sample ConfigMap and Metric
One particular place this could be used is to track automated server rebuilding in your cluster and have the automation pipeline update the ConfigMap along the way for the server it is currently building.

When this ConfigMap is applied a metric would get registered for node_repave_status, the metric would have a label "hostame", with a value of "node1" and the metric value would be 1.
```
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: c2m-node-repave-phase-node1
  namespace: default
  labels:
    prom_metric: node_repave_phase
    prom_description: "The_repave_phase"
    prom_labels: "hostname"
data:
  node1: "1"
```

```
# HELP node_repave_phase The repave phase.
# TYPE node_repave_phase gauge
node_repave_phase{hostname="node1"} 1
```

As the process is happening you would update the metric with different values, here are some samples that follow this specific process.

*  0 means server is up and running
*  1 means server is starting a drain.
*  2 means server is starting a rebuild.
*  3 means server is getting prepared after rebuild.
*  4 means server is getting joined to the cluster.
*  5 means server is getting post join configs.

SLA's can be measured against any specifc part of the process so alerts can be generated in a process is in a phase for to long.