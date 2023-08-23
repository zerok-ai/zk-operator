#!/bin/bash
THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
echo "THIS_DIR=$THIS_DIR"
echo "DOCKER_REPO=$DOCKER_REPO"
echo "ZK_INIT_REPO=$ZK_INIT_REPO"
echo "ZK_INIT_TAG=$ZK_INIT_TAG"

# Set the image tag based on the chart version
perl -pi -e "s#dockerBase: \".*\"#dockerBase: \"$DOCKER_REPO\"#" $THIS_DIR/values.yaml
perl -pi -e "s#zkInitContainerRepo: \".*\"#dockerBase: \"$ZK_INIT_REPO\"#" $THIS_DIR/values.yaml
perl -pi -e "s#zkInitContainerTag: \".*\"#dockerBase: \"$ZK_INIT_TAG\"#" $THIS_DIR/values.yaml