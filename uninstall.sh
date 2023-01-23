#!/bin/bash

#TODO: Create a new service account and only give access to minimum needed actions on the cluster.
scriptDir=$(dirname -- "$(readlink -f -- "$BASH_SOURCE")")

kubectl delete clusterrolebinding serviceaccounts-cluster-admin \
  --clusterrole=cluster-admin \
  --group=system:serviceaccounts

make -C ${scriptDir} undeploy
