#!/bin/bash

#TODO: Create a new service account and only give access to minimum needed actions on the cluster.
scriptDir=$(dirname -- "$(readlink -f -- "$BASH_SOURCE")")

kubectl create clusterrolebinding serviceaccounts-cluster-admin \
  --clusterrole=cluster-admin \
  --group=system:serviceaccounts

make -C ${scriptDir} generate
make -C ${scriptDir} manifests
if [ "$1" = "build" ]; then
  make -C ${scriptDir} gke docker-build docker-push
fi
make -C ${scriptDir} deploy

