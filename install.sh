#TODO: Create a new service account and only give access to minimum needed actions on the cluster.
kubectl create clusterrolebinding serviceaccounts-cluster-admin \
  --clusterrole=cluster-admin \
  --group=system:serviceaccounts
make generate
make manifests
if [ "$1" = "build" ]; then
    make docker-build docker-push
fi
make deploy
kubectl delete -f config/samples/operator_v1alpha1_zerokop.yaml
kubectl apply -f config/samples/operator_v1alpha1_zerokop.yaml