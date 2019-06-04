GOLANG_VERSION := 1.12.5
CODE_GENERATOR_VERSION := 1.14.2
KUBE_OPENAPI_VERSION := a01b7d5d6c2258c80a4a10070f3dee9cd575d9c7
GOPKG = github.com/s-urbaniak/prometheus-adapter
REPO ?= quay.io/surbania/code-generator
GEN_ARGS = --v=1 --logtostderr --go-header-file scripts/boilerplate.txt

REGISTER_TARGET:=pkg/apis/metrics/v1/register_generated.go
$(REGISTER_TARGET):
	register-gen \
	$(GEN_ARGS) \
	--input-dirs "$(GOPKG)/pkg/apis/metrics/v1" \
	--output-file-base register_generated

DEEPCOPY_TARGET := pkg/apis/metrics/v1/deepcopy_generated.go
$(DEEPCOPY_TARGET): $(REGISTER_TARGET)
	deepcopy-gen \
	$(GEN_ARGS) \
	--input-dirs      "$(GOPKG)/pkg/apis/metrics/v1" \
	--bounding-dirs   "$(GOPKG)/pkg/apis/metrics" \
	--output-file-base deepcopy_generated

CLIENT_TARGET := pkg/client/clientset/versioned/clientset.go
$(CLIENT_TARGET): $(DEEPCOPY_TARGET)
	client-gen \
	$(GEN_ARGS) \
	--clientset-name "versioned" \
	--input-base     "" \
	--input          $(GOPKG)/pkg/apis/metrics/v1 \
	--clientset-path $(GOPKG)/pkg/client/clientset

LISTER_TARGET := pkg/client/listers/metrics/v1/resourcerule.go
$(LISTER_TARGET):
	lister-gen \
	$(GEN_ARGS) \
	--input-dirs     "$(GOPKG)/pkg/apis/metrics/v1" \
	--output-package "$(GOPKG)/pkg/client/listers"

INFORMER_TARGET := pkg/client/informers/externalversions/metrics/interface.go
$(INFORMER_TARGET): $(LISTER_TARGET) $(CLIENT_TARGET)
	informer-gen \
	$(GEN_ARGS) \
	--input-dirs                  "$(GOPKG)/pkg/apis/metrics/v1" \
	--versioned-clientset-package "$(GOPKG)/pkg/client/clientset/versioned" \
	--listers-package             "$(GOPKG)/pkg/client/listers" \
	--output-package              "$(GOPKG)/pkg/client/informers"

OPENAPI_TARGET := pkg/apis/metrics/v1/openapi_generated.go
$(OPENAPI_TARGET):
	openapi-gen \
	$(GEN_ARGS) \
	--input-dirs $(GOPKG)/pkg/apis/metrics/v1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/api/core/v1 \
	--output-package $(GOPKG)/pkg/apis/metrics/v1 \
	--output-file-base openapi_generated

.PHONY: build-image
build-image:
	docker build \
	--build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
	--build-arg CODE_GENERATOR_VERSION=kubernetes-$(CODE_GENERATOR_VERSION) \
	--build-arg KUBE_OPENAPI_VERSION=$(KUBE_OPENAPI_VERSION) \
	-f Dockerfile.build \
	-t $(REPO):v$(CODE_GENERATOR_VERSION) \
	.

.PHONY: all
all: \
	$(REGISTER_TARGET) \
	$(DEEPCOPY_TARGET) \
	$(CLIENT_TARGET) \
	$(LISTER_TARGET) \
	$(INFORMER_TARGET) \
	$(OPENAPI_TARGET)

.PHONY: clean
clean:
	rm -rf pkg/client
	rm -f \
		$(DEEPCOPY_TARGET) \
		$(REGISTER_TARGET) \
		$(OPENAPI_TARGET)
