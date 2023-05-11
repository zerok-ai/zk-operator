#!/bin/bash

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $THIS_DIR/variables.sh

kubectl delete mutatingwebhookconfiguration zk-webhook
make -C ${ROOT_DIR} undeploy
