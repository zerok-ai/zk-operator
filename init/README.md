# init-container 
The init container bundles the opentelemetry agent and extensions and mounts them to a path in the target container.

# Pre-requisites
The init container will be added as a patch to the target container by the zerok operator. When updating the init container, you should also update the operator to pick the right version.

### Building
1. To create a new build of the init container and push to gke.

```
make buildAndPush
```