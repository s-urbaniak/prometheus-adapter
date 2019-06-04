module github.com/s-urbaniak/prometheus-adapter

go 1.12

replace (
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v0.0.0-20190525122359-d20e84d0fb64
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible
)

require (
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/kubernetes-incubator/metrics-server v0.3.3
	github.com/prometheus/client_golang v0.9.3
	github.com/prometheus/common v0.4.1
	github.com/prometheus/prometheus v0.0.0-00010101000000-000000000000
	k8s.io/api v0.0.0-20190602205700-9b8cae951d65
	k8s.io/apimachinery v0.0.0-20190602113612-63a6072eb563
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/metrics v0.0.0-20190531135401-156151eebb71
)
