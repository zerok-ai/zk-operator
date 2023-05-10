# Zerok-operator
Zerok operator in responsible for instrumenting the pods coming up in the cluster. It achieves this by adding a mutatingadmissionwebhook in the cluster.

# Pre-requisites
It needs Redis to be up and running as it uses Redis to read language data for images. This data will be populated by the zerok-deamonset pod. The operator pod will sync the data from Redis based on the time interval specified in the config file.

### Quickstart

1. To create a new build and install the controller on the cluster.
```
./buildAndInstall.sh
```

### Running on the cluster
1. Install the operator on a running cluster.

```
./install.sh
```

2. Build and push your image. The variables for creating the file are present in the Makefile.
	
```sh
make docker-build docker-push
```

### Testing
To test the webhook request coming from the control plane, you can test using the Webhook Request present in the zk-operator postman collection. 