#!/bin/bash

scriptDir=$(dirname -- "$(readlink -f -- "$BASH_SOURCE")")

kubectl delete mutatingwebhookconfiguration zerok-webhook
kubectl delete namespace zerok-injector
make -C ${scriptDir} undeploy
kubectl delete -f $scriptDir/config/samples/operator_v1alpha1_zerokop.yaml
