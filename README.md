# kdmq (kdm query)

Tool to query KDM data for a given Rancher version, think of:

- What k8s version are included in a Rancher release?
- What k8s images are in a Rancher Kubernetes version?
- What addons are included in a Rancher Kubernetes version?

And diff commands to automatically get the differences between two versions.

## Docs

Under construction!

Parameters:

- `rancher_version`: The version of Rancher that is used for querying KDM data, examples: v2.5.10, v2.6.3, v2.6.0
- `channel`: The source of KDM data used, valid options are:
  - `release` (what is in the released version, no out-of-band data)
  - `latest` (what is currently released in KDM and active to installs refreshing KDM from online source, coming from https://releases.rancher.com/kontainer-driver-metadata/release-(v2.5|v2.6)/data.json)
  - `dev` (what is currently in development, coming from https://releases.rancher.com/kontainer-driver-metadata/dev-(v2.5|v2.6)/data.json))
  - `./$FILE` (local data file, must be prefixed with `./` to indicate local data file)
  - `https://URL_TO/data.json` (remote URL data file, must be a valid URL to be used)


Flags:

- `verbose`: Print every piece of data used.
- `diff-oneway`: Only diff one way (defaults to two way)
- `show-all`: Show all k8s versions available (only for RKE2/K3S for now)
- `output`: Option to print output as json (`--output=json`)

## Examples

* List k8s versions for a Rancher version

```
kdmq listk8s rke v2.6.3 release
kdmq listk8s rke2 v2.6.3 latest
kdmq --show-all listk8s rke2 v2.6.3 latest // only for RKE2/K3S
kdmq listk8s k3s v2.6.3 dev
kdmq --show-all listk8s k3s v2.6.3 dev // only for RKE2/K3S
```

* List k8s versions for a Rancher version using a remote URL data file

```
kdmq listk8s rke v2.6.3 https://my.domain.com/data.json
```

* Diff k8s versions for Rancher versions

```
kdmq diffk8s rke2 v2.6.3 v2.6.3 release latest
```

* Diff oneway k8s versions for Rancher versions

```
kdmq --diff-oneway diffk8s rke v2.6.3 v2.6.3 release latest
```

* Diff verbose k8s versions for Rancher versions

```
kdmq --verbose diffk8s rke v2.6.3 v2.6.3 release latest
```

* List k8s images for a Rancher k8s version

```
kdmq listk8simages v1.22.5-rancher1-1 v2.6.3 latest
```

* Diff k8s images for Rancher k8s versions

```
kdmq diffk8simages v1.22.5-rancher1-1 v1.22.4-rancher1-1 v2.6.3 dev latest
```

* Diff k8s images for Rancher k8s versions with local data

```
kdmq diffk8simages v1.22.5-rancher1-1 v1.22.4-rancher1-1 v2.6.3 dev ./data.json
```

* Diff oneway all k8s images between Rancher k8s versions between Rancher version

```
kdmq --diff-oneway diffallk8simages v2.6.0 v2.6.3 release release
```

* List k8s addons for Rancher k8s version

```
kdmq listk8saddons v1.22.5-rancher1-1 v2.6.3 latest
```

* Diff k8s addons for Rancher k8s versions

```
kdmq diffk8saddons v1.22.5-rancher1-1 v1.21.7-rancher1-1 v2.6.3 latest latest
```
