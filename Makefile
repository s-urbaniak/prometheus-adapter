GOLANG_VERSION=1.12.5
CODE_GENERATOR_VERSION=kubernetes-1.14.2
KUBE_OPENAPI_VERSION=a01b7d5d6c2258c80a4a10070f3dee9cd575d9c7

GOPKG=github.com/s-urbaniak/prometheus-adapter
GROUP=metrics.prometheus.io

GEN_ARGS=--v=1 --logtostderr

pkg/apis/$(GROUP)/v1/zz_generated.deepcopy.go:
	deepcopy-gen \
	$(GEN_ARGS) \
	--input-dirs    "$(GOPKG)/pkg/apis/$(GROUP)/v1" \
	--bounding-dirs "$(GOPKG)/pkg/apis/$(GROUP)" \
	--output-file-base zz_generated.deepcopy \
	--go-header-file .header

pkg/client/clientset/versioned/clientset.go:
	client-gen \
	$(GEN_ARGS) \
	--clientset-name "versioned" \
	--input-base "" \
	--input $(GOPKG)/pkg/apis/$(GROUP)/v1 \
	--clientset-path $(GOPKG)/pkg/client/clientset \
	--go-header-file .header

.PHONY: build-image
build-image:
	docker build \
	--build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
	--build-arg CODE_GENERATOR_VERSION=$(CODE_GENERATOR_VERSION) \
	--build-arg KUBE_OPENAPI_VERSION=$(KUBE_OPENAPI_VERSION) \
	-f Dockerfile.build \
	-t quay.io/surbania/code-generator:latest \
	.

.PHONY: all
all: \
	pkg/apis/$(GROUP)/v1/zz_generated.deepcopy.go \
	pkg/client/clientset/versioned/clientset.go

.PHONY: clean
clean:
	rm -rf pkg/client
	rm -f pkg/apis/$(GROUP)/v1/zz_generated.deepcopy.go
