# zerok-operator
Zerok operator sets up the required zerok components on the cluster. 

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