# Zerok-operator
Zerok operator in responsible for instrumenting the pods coming up in the cluster. It achieves this by adding a mutatingadmissionwebhook in the cluster.

# Pre-requisites
It needs Redis to be up and running as it uses Redis to read language data for images. This data will be populated by the zerok-deamonset pod. The operator pod will sync the data from Redis based on the time interval specified in the config file.


### Running on the cluster
1. Install the operator on a running cluster without rebuild.

```
make install
```

2. To create a new build of the operator and push to gke.

```
make build
```

3. Uninstall operator

```
make uninstall
```