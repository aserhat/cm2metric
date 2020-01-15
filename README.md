# cm2mertics

A quick proof of concept to use a ConfigMap to hold metric information and a small app to continually read the ConfigMap and expose an endpoint serving the metrics.  This could be used in a place for where is not easy to expose a metric but you can make an update to the ConfigMap easier.  Prometheus could scrape that metric and make it availble for dashboards and alerting.  One particular place this is used is to track automated server repaving.

Current Propsed Mertic

*  0 means server is up.
*  1 means server is starting a repave.
*  2 means server is starting a rebuild.
