# Prometheus Adapter Overhaul

## Summary

This document proposes the motivation of the rewrite of prometheus adapter, the envisioned user experience, and the target architecture.

## Motivation

Currently, [prometheus adapter](https://github.com/DirectXMan12/k8s-prometheus-adapter) fills the gap for all three Kubernetes APIs (resource metrics, custom metrics, external metrics).

The current implementation of prometheus adapter has a few problems which need to be addressed:

- Problematic metrics discovery mechanism having known scalability issues
- Complex configuration
- Static, central configuration for all exposed Kubernetes metrics

### Goals

* Improve scalability
* Improve user experience
* Lower the configuration complexity of the existing solution
* Allow dynamic provisioning of new custom/external metrics without restarts
* Host the prometheus adapter source code in the https://github.com/kubernetes-sigs organization
* RBAC awareness

### Non-Goals

* Multi-tenancy

## Proposal

### Configuration

#### Resource Metrics

The resource metrics API in Kubernetes consists of predefined metrics, hard-coded in the API.
The following two metrics are considered to be mapped for the prometheus adapter:

- cpu metrics
- memory metrics

The metrics above are available throughout the following dimensions:

- Node (non-namespaced)
- Pods (namespaced)

This configuration is considered to be static and not to changed throughout the lifetime of a cluster, mostly.
For this reason it is proposed that this is simply continued to be configured using a configmap,
having the following structure:

```
rules:
  cpu:
    containerLabel: container_name
    labels:
      namespace:
        resource: namespace
      node:
        resource: node
      pod_name:
        resource: pod
    queries:
      node: sum(1 - rate(node_cpu_seconds_total{mode="idle"}[1m]) * on(namespace,
        pod) group_left(node) node_namespace_pod:kube_pod_info:{node="kind-control-plane"})
        by (node)
      pod: sum(rate(container_cpu_usage_seconds_total{container_name!="POD",container_name!="",pod_name="alertmanager-main-0",pod_name!="",namespace="monitoring"}[1m]))
        by (container_name,pod_name,namespace)
  memory:
    containerLabel: container_name
    labels:
      namespace:
        resource: namespace
      node:
        resource: node
      pod_name:
        resource: pod
    queries:
      node: sum(node:node_memory_bytes_total:sum{node="kind-control-plane"} - node:node_memory_bytes_available:sum{node="kind-control-plane"})
        by (node)
      pod: sum(container_memory_working_set_bytes{pod_name="alertmanager-main-0",namespace="monitoring",container_name!="POD",container_name!="",pod_name!=""})
        by (pod_name,namespace,container_name)
window: 0s
```

#### Custom Metrics

TBD - copy and explain CRD definition from the PoC

#### External Metrics

TBD (as above)

### Dynamic metrics provisioning

This point is simply solved by the nature of CRDs. By moving to CRDs, prometheus adapter acts like yet another controller, watching changes on any custom and external metrics custom resource. Upon change, it updates its internal registry.

### Scalability

The major issue of the current implementation of prometheus adapter is its metrics discoverability mechanism which can easily lead to out of memory errors when the discoverability query ingests too many metrics series.

The proposed solution is to completely omit automatic discoverability of metrics. Similar to Zalando's [kube-metrics-adapter](https://github.com/zalando-incubator/kube-metrics-adapter) we believe that the better operational model is to pre-declare metrics explicitely.

## Upgrade / Downgrade Strategy

