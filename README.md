# Cluster Version Operator plugin

## What?
This application is a plugin for `kubectl`/`oc` when using OpenShift clusters.
It's goal is to ease `ClusterOperator` development, as these are managed by
CVO (cluster version operator), and all changes made there are immediately
overwritten.

## Why?
To be able to change the operator's `Deployment` spec, we have to set up
special overrides in the `ClusterVersion` resource. While this is not complex,
it might get cumbersome to remember the patch or edit the resource, so this
plugin eases that out for you. On top of that, taking advantage of this being
a plugin, it allows managing and unmanaging operators, listing the available
options based on current status.

## How?
The plugin is intended for human usage and is therefore interactive (a
non-interactive version may be added if needed). Once downloaded, you can
build and install.
### Build
Easy, simply execute:
```bash
$ make build
```

### Install
Easy again, simply execute:
```bash
$ make install-plugin
```

You will see a new directory is created: `$HOME/bin`. This is what `kubectl`/`oc`
plugins use as one of the locations to look them up. Once installed you can execute
`kubectl cvo` or `oc cvo`.

### Usage
The plugin provides two commands:
```bash
$ oc cvo --help
This application interacts with OCP clusters to act upon CVO tasks

Usage:
  cvo [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  manage      Removes an override in CVO (if it exists) to manage an operator
  unmanage    Sets an override in CVO to unmanage an operator

Flags:
  -h, --help   help for cvo

Use "cvo [command] --help" for more information about a command.
```

Both `manage` and `unmanage` commands are interactive. `Manage` removes
overrides in a `ClusterVersion` resource to get an operator back to a managed
state. `Unmanage` does the opposite. Both commands fetch information from the
cluster to show current managed/unmanaged operators to select them from a list:

```bash
$ oc cvo unmanage
? Choose an operator:  [Use arrows to move, type to filter]
> version:openshift-operator-lifecycle-manager/package-server-manager
  version:openshift-cloud-credential-operator/cloud-credential-operator
  version:openshift-cluster-node-tuning-operator/cluster-node-tuning-operator
  version:openshift-controller-manager-operator/openshift-controller-manager-operator
  version:openshift-operator-lifecycle-manager/olm-operator
  version:openshift-machine-api/cluster-autoscaler-operator
  version:openshift-machine-api/machine-api-operator
```
What you see follows this naming convention: `<cluster version resource name>:<namespace>/<operator name>`.

The list selection is also interactive, you can start typing to filter operator
names:
```bash
$ oc cvo unmanage
? Choose an operator: ing  [Use arrows to move, type to filter]
> version:openshift-cluster-node-tuning-operator/cluster-node-tuning-operator
  version:openshift-monitoring/cluster-monitoring-operator
  version:openshift-ingress-operator/ingress-operator
```
### Example
Let's unmanage an operator, check the result, and manage it again:
```bash
# Choose the cluster-monitoring-operator from the list, this is how it looks like after selection.
$ oc cvo unmanage
? Choose an operator: version:openshift-monitoring/cluster-monitoring-operator

# Check the overrides present in `version` clusterversion.
$ oc get clusterversion version -o jsonpath='{.spec.overrides}' | jq
{
  "channel": "stable-4.10",
  "clusterID": "992fd4f4-de71-4006-86bb-840d1ef26379",
  "overrides": [
    {
      "group": "apps",
      "kind": "Deployment",
      "name": "cluster-monitoring-operator",
      "namespace": "openshift-monitoring",
      "unmanaged": true
    }
  ]
}

# Now we execute the manage command, At first we see this, as there is only one unmanaged operator available
# for return to managed again
$ oc cvo manage
? Choose an operator:  [Use arrows to move, type to filter]
> version:openshift-monitoring/cluster-monitoring-operator

# Now we execute it by hitting enter
$ oc cvo manage
? Choose an operator: version:openshift-monitoring/cluster-monitoring-operator

# And now we check the overrides again
$ oc get clusterversion version -o jsonpath='{.spec}' | jq
{
  "channel": "stable-4.10",
  "clusterID": "992fd4f4-de71-4006-86bb-840d1ef26379"
}
```