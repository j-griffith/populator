# Kubernetes Populator Custom Resource Definition (CRD) and Controller

## Clone it

`git clone https://github.com/j-griffith/populator $GOPATH/src/github.com/j-griffith/populator`

## Build it

`make all`

## Install the CRD

`make install`

## Launch the controller/manager

TBD: We'll have ``make deploy`` BUT that doesn't work right now!  So we just run locally (only working with local-up-cluster.sh right now)

`bin/populator-controller -kubeconfig $HOME/.kube/config`

## Create a Populator object

`kubectl create -f kubernetes/populator.yaml`

## Create a PVC the uses the Populator

`kubectl create -f kubernetes/pvc-populator-src.yaml`
