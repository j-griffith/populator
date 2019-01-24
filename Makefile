IMG ?= populator-controller:latest

all: manager
	go build ./pkg/api/types/v1alpha1/
	go build ./pkg/clientset/v1alpha1/
	go build ./pkg/controller/
	go build ./pkg/populator/

# Build the manager (controller) binary
manager:
	go build -o bin/populator-controller github.com/j-griffith/populator/cmd/manager

# Install the Populator CRD to the cluster
install: 
	kubectl apply -f kubernetes/crd.yaml

# TODO: add a deploy that will deploy the controller for us

docker-build:

docker-push:

