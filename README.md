# Prometheus Adapter for Kubernetes Metrics APIs

This repository contains an implementation of the Kubernetes
[resource metrics](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/instrumentation/resource-metrics-api.md) API and
[custom metrics](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/instrumentation/custom-metrics-api.md) API.

This adapter is therefore suitable for use with the autoscaling/v2 Horizontal Pod Autoscaler in Kubernetes 1.14+.
It can also replace the [metrics server](https://github.com/kubernetes-incubator/metrics-server) on clusters that already run Prometheus and collect the appropriate metrics.

## FAQ

### Bumping k8s dependencies

1. Bump the corresponding versions in go.mod
```
$ cat go.mod
...
	k8s.io/api kubernetes-1.14.2
	k8s.io/apimachinery kubernetes-1.14.2
	k8s.io/client-go v11.0.0
...
```

2. Execute:
```
$ go mod vendor
```
