GOLANG_VERSION:=1.12.5
CODE_GENERATOR_VERSION:=kubernetes-1.14.2
KUBE_OPENAPI_VERSION:=a01b7d5d6c2258c80a4a10070f3dee9cd575d9c7

GOPKG=github.com/s-urbaniak/prometheus-adapter
GROUP=metrics

GEN_ARGS=--v=1 --logtostderr --go-header-file .header

DEEPCOPY_TARGET:=pkg/apis/$(GROUP)/v1/zz_generated.deepcopy.go
$(DEEPCOPY_TARGET):
	deepcopy-gen \
	$(GEN_ARGS) \
	--input-dirs      "$(GOPKG)/pkg/apis/metrics/v1" \
	--bounding-dirs   "$(GOPKG)/pkg/apis/metrics" \
	--output-file-base zz_generated.deepcopy

CLIENT_TARGET:=pkg/client/clientset/versioned/clientset.go
$(CLIENT_TARGET):
	client-gen \
	$(GEN_ARGS) \
	--clientset-name "versioned" \
	--input-base     "" \
	--input          $(GOPKG)/pkg/apis/metrics/v1 \
	--clientset-path $(GOPKG)/pkg/client/clientset

LISTER_TARGET:=pkg/client/listers/$(GROUP)/v1/resourcerule.go
$(LISTER_TARGET):
	lister-gen \
	$(GEN_ARGS) \
	--input-dirs     "$(GOPKG)/pkg/apis/metrics/v1" \
	--output-package "$(GOPKG)/pkg/client/listers"

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
	$(DEEPCOPY_TARGET) \
	$(CLIENT_TARGET) \
	$(LISTER_TARGET)

.PHONY: clean
clean:
	rm -rf pkg/client
	rm -f pkg/apis/metrics/v1/zz_generated.deepcopy.go
