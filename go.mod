module github.com/s-urbaniak/prometheus-adapter

go 1.12

require (
	github.com/kubernetes-incubator/metrics-server v0.3.3
	github.com/prometheus/prometheus v0.0.0-20190525122359-d20e84d0fb64
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	google.golang.org/appengine v1.5.0 // indirect
	k8s.io/api v0.0.0-20190528154508-67ef80593b24
	k8s.io/apimachinery v0.0.0-20190528154326-e59c2fb0a8e5
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/metrics v0.0.0-20190528160841-39fcca00df64
)
