# Zerok-operator
Using the Zerok Operator, traces can be filtered using probes to enhance observability within Kubernetes clusters. A probe is a set of rules defined by the user for capturing traces of interest. The probes are created using the `ZerokProbe` custom resource definition (CRD). You can refer to the [ZEROKPROBE.md](ZEROKPROBE.md) for details about creating the `ZerokProbe` CRD. 

# Pre-requisites
An operational Redis instance is required within the cluster or accessible externally, used by the operator to store probe rules information.

<Add info for design and installation steps here>.

### Contributing
Contributions to the Zerok Operator are welcome! Submit bug reports, feature requests, or code contributions.

### Reporting Issues
Encounter an issue? Please file a report on our GitHub issues page with detailed information to aid in quick resolution.