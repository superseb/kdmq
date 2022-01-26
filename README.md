# kdmq (kdm query)

Tool to query KDM data for a given Rancher version, think of:

- What k8s version are included in a Rancher release (channel `embedded`), out of band (channel `release`) or in development (channel `dev`)?
- What k8s images are in a Rancher Kubernetes version?
- What addons are included in a Rancher Kubernetes version?

And diff commands to automatically get the differences between two version.

## Docs

Under construction!

Parameters:

- `rancher_version`: The version of Rancher that is used for querying KDM data, examples: v2.5.10, v2.6.3, v2.6.0
- `channel`: The source of KDM data used, valid options are: `embedded` (what is in the released version, no out-of-band data), `release` (what is currently released in KDM and active to installs refreshing KDM from online source), `dev` (what is currently in development), `./$FILE` (local data file, must be prefixed with `./` to indicate local data file)

## Examples

* List k8s versions for a Rancher version

```
kdmq listk8s v2.6.3 release
```

* Diff k8s versions for Rancher versions

```
kdmq diffk8s v2.6.3 v2.6.3 embedded release
```

* Diff oneway k8s versions for Rancher versions

```
kdmq --diff-oneway diffk8s v2.6.3 v2.6.3 embedded release
```

* Diff verbose k8s versions for Rancher versions

```
kdmq --verbose diffk8s v2.6.3 v2.6.3 embedded release
```

* List k8s images for a Rancher k8s version

```
kdmq listk8simages v1.22.5-rancher1-1 v2.6.3 release
```

* Diff k8s images for Rancher k8s versions

```
kdmq diffk8simages v1.22.5-rancher1-1 v1.22.4-rancher1-1 v2.6.3 dev release
```

* Diff k8s images for Rancher k8s versions with local data

```
kdmq diffk8simages v1.22.5-rancher1-1 v1.22.4-rancher1-1 v2.6.3 dev ./data.json
```

* Diff oneway all k8s images between Rancher k8s versions between Rancher version

```
kdmq --diff-oneway diffallk8simages v2.6.0 v2.6.3 embedded embedded
```

* List k8s addons for Rancher k8s version

```
kdmq listk8saddons v1.22.5-rancher1-1 v2.6.3 release
```

* Diff k8s addons for Rancher k8s versions

```
kdmq diffk8saddons v1.22.5-rancher1-1 v1.21.7-rancher1-1 v2.6.3 release release
```
