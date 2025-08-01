# SPDX-FileCopyrightText: the secureCodeBox authors
#
# SPDX-License-Identifier: Apache-2.0

include ../../../prerequisites.mk

IMG_NS ?= securecodebox

# Image URL to use all building/pushing image targets
IMG ?= auto-discovery-secret-extractor

# Tag used for the image
IMG_TAG ?= sha-$$(git rev-parse --short HEAD)


##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##@ Development

.PHONY: test
test: unit-test integration-test

.PHONY: unit-test
unit-test:
	$(PYTHON) -m unittest discover

.PHONY: integration-test
integration-test: docker-build docker-export kind-import
	@echo ".: 🩺 Starting integration test in kind namespace 'integration-tests'."
	kubectl delete namespace integration-test --wait || true
	kubectl create namespace integration-test
	
	./integration-test/test-pod.sh ${IMG_NS}/${IMG}:${IMG_TAG}
	kubectl wait --for=condition=ready --timeout=60s -n integration-test pod/init-container-test

	kubectl get secret --namespace integration-test test-secret

##@ Build


.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	@echo ".: ⚙️ Build Container Images"
	docker build -t ${IMG_NS}/${IMG}:${IMG_TAG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG_NS}/${IMG}:${IMG_TAG}

.PHONY: docker-export
docker-export:
	@echo ".: 💾 Export Container Images"
	docker save $(IMG_NS)/$(IMG):$(IMG_TAG) > $(IMG).tar

##@ Deployment

.PHONY: kind-import
kind-import:
	@echo ".: 💾 Importing the image archive to local kind cluster."
	kind load image-archive ./$(IMG).tar