# Zerok-operator
The Zerok Operator is part of the Zerok system [add link here], which is a set of tools for observability in Kubernetes clusters. The Zerok system works along with the Opentelemetry Operator. Check out these docs [add link here] to learn more about how Zerok can benefit you. 

The Zerok Operator is a Kubernetes operator that provides a custom resource definition (CRD) for creating probes to capture traces of interest within Kubernetes clusters. A probe is a set of rules defined by the user for capturing traces of interest. The probes are created using the `ZerokProbe` custom resource definition (CRD). You can refer to the [ZEROKPROBE.md](ZEROKPROBE.md) for details about creating the `ZerokProbe` CRD. 

## Prerequisites
Redis needs to be installed in the cluster in zk-client namespace for the operator to work. Zerok Operator uses Redis as a backend to store the probe data. Please refer to the steps below for setting up Redis and the operator.


## Get Helm Repositories Info

```console
helm repo add zerok-ai https://zerok-ai.github.io/helm-charts
helm repo update
```

_See [`helm repo`](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install Helm Chart

Install redis in zk-client namespace. This optional step is required only if redis is not installed in the cluster.
```console
helm install [RELEASE_NAME] zerok-ai/zk-redis
```

Install the Zerok Operator.
```console
helm install [RELEASE_NAME] zerok-ai/zk-operator
```

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Helm Chart

```console
helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Helm Chart

```console
helm upgrade [RELEASE_NAME] [CHART] --install
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

### Contributing
Contributions to the Zerok Operator are welcome! Submit bug reports, feature requests, or code contributions.

### Reporting Issues
Encounter an issue? Please file a report on our GitHub issues page with detailed information to aid in quick resolution.