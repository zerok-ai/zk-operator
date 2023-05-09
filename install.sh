#!/bin/bash

SCRIPTDIR=$(dirname -- "$(readlink -f -- "$BASH_SOURCE")")

IMG_BASE=us-west1-docker.pkg.dev/zerok-dev/stage/zerok-operator
VERSION=test
IMG=$IMG_BASE:$VERSION

LOCALBIN=$SCRIPTDIR/bin
CONTROLLER_GEN=$LOCALBIN/controller-gen
KUSTOMIZE=$LOCALBIN/kustomize

CONTROLLER_TOOLS_VERSION=v0.9.2


KUSTOMIZE_INSTALL_SCRIPT="https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
test -s $LOCALBIN/kustomize || curl -s $KUSTOMIZE_INSTALL_SCRIPT | bash -s $LOCALBIN
test -s $LOCALBIN/controller-gen || GOBIN=$LOCALBIN go install sigs.k8s.io/controller-tools/cmd/controller-gen@$CONTROLLER_TOOLS_VERSION


$CONTROLLER_GEN rbac:roleName=manager-role crd webhook paths="$SCRIPTDIR/..." output:crd:artifacts:config=config/crd/bases
cd $SCRIPTDIR/config/manager && $KUSTOMIZE edit set image controller=$IMG
$KUSTOMIZE build $SCRIPTDIR/config/default | kubectl apply -f -
kubectl apply -f $SCRIPTDIR/config/samples/operator_v1alpha1_zerokop.yaml