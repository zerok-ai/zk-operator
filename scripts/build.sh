#!/bin/bash

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $THIS_DIR/variables.sh

make -C ${ROOT_DIR} generate
make -C ${ROOT_DIR} manifests
make -C ${ROOT_DIR} gke docker-build docker-push