# Zerok-operator
Using the Zerok Operator, traces can be filtered using probes to enhance observability within Kubernetes clusters. A probe is a set of rules defined by the user for capturing traces of interest. The probes are created using the `ZerokInstrumentation` custom resource definition (CRD). You can refer to the `doc` for details about creating the `ZerokInstrumentation` CRD. 

# Pre-requisites
An operational Redis instance is required within the cluster or accessible externally, used by the operator to store probe information.

### Running on the cluster
1. Install the operator on a running cluster without rebuild.

```
make install
```

2. To create a new build of the operator and push to gke.

```
make buildAndPush
```

3. Uninstall operator

```
make uninstall
```

### Contributing
Contributions to the Zerok Operator are welcome! Submit bug reports, feature requests, or code contributions.

### Reporting Issues
Encounter an issue? Please file a report on our GitHub issues page with detailed information to aid in quick resolution.