# cm2metrics

[![](https://github.com/aserhat/cm2metric/workflows/Build/badge.svg)](https://github.com/aserhat/cm2metric/actions)
[![Go Report Card](https://goreportcard.com/badge/aserhat/cm2metric)](https://goreportcard.com/report/aserhat/cm2metric)
[![codecov](https://codecov.io/gh/aserhat/cm2metric/branch/main/graph/badge.svg)](https://codecov.io/gh/aserhat/cm2metric)
[![Releases](https://img.shields.io/github/release-pre/aserhat/cm2metric.svg?sort=semver)](https://github.com/aserhat/cm2metric/releases)
[![LICENSE](https://img.shields.io/github/license/aserhat/cm2metric.svg)](https://github.com/aserhat/cm2metric/blob/master/LICENSE)


An application that use ConfigMaps as the source of metric information, reads them through the use of SharedInformers and exports the data as Prometheus metrics.  The main use of this application is to easily export metrics for processes like automation pipelines that might not easily have a way of doing so. 

## Sample ConfigMap and Metric
An example here is a pipeline that performas automated server rebuilding.  Through each step of the pipeline the ConfigMap is updated and the value of the data is updated to represent a change in the metric.

When this ConfigMap is create for the first time a metric gets created and registered for node_rebuild_phase, the metric would have a label "hostname", with a value of "node1" and the metric value would be 1.
```
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: c2m-node-rebuild-phase-node1
  namespace: default
  labels:
    prom_metric: node_rebuild_phase
    prom_description: "The_server_rebuild_phase"
    prom_labels: "hostname"
data:
  node1: "1"
```

```
# HELP node_rebuild_phase The server rebuild phase.
# TYPE node_rebuild_phase gauge
node_rebuild_phase{hostname="node1"} 1
```

As the pipeline is running it would update the ConfigMap/metric with different values, here are some samples that follow this specific process.

*  0 means server is up and running
*  1 means server is starting a drain.
*  2 means server is starting a rebuild.
*  3 means server is getting prepared after rebuild.
*  4 means server is getting joined to the cluster.
*  5 means server is getting post join configs.

SLA's can be measured against any specifc part of the process so alerts can be generated if a process is in a phase for to long.
