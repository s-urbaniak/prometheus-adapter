## Bumping k8s dependencies

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
