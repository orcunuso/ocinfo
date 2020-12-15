![OCinfo](https://github.com/vorcunus/ocinfo/blob/main/png/ocinfo.png?raw=true)

OCinfo is a tool written in pure Go that was influenced from the hassles of managing multiple OpenShift clusters and the need to improve visibility. What it simple does is to get data from OpenShift APIs with the readonly credentials provided, prints out all the data in a pretty, human-readable and analyzable Microsoft Excel &trade; spreadsheet document and save it as a local xlsx file. With Go, this can be done with an independent binary distribution across all platforms that Go supports, including Linux, MacOS, Windows and ARM.

OCinfo runs best with OpenShift 4.4 and 4.5, support for new features of OpenShift 4.6 like OVNKubernetes is on the roadmap. 

## Installation

If you have an installed go environment, you can get source code and compile.

```bash
go get github.com/vorcunus/ocinfo
cd $GOPATH/src/github.com/vorcunus/ocinfo
make compile_all
```

If you don't have an installed go environment, you can also download compiled binaries from [releases](https://github.com/vOrcunus/ocinfo/releases) (coming soon).

## Preperation

There are a few things that need to be configured or checked before running OCinfo;

1. Make sure that the platform running OCinfo has network access to all OpenShift APIs
2. A YAML file is created to configure OCinfo
3. A service account token from all OpenShift clusters

### OCinfo configuration

OCinfo expects a YAML file to be given as an input to get more information about the OpenShift clusters that it will connect to and some configuration details related with spreadsheet document as the output. OCinfo will look for a file named `ocinfo.yaml` in working directory but this behavior can be altered with flag `-f`.

An example file exists in `./conf` directory and it looks like this;

```yaml
clusters:
- name: cluster1
  enable: true
  baseURL: "https://api.cluster1.orcunuso.io:6443"
  token: "eyJhbGciOiJSUzI1.........YKWA8UfWzLsuwjfEPphPPlwa7SA"
  quota: "quota-compute"
- name: cluster2
  enable: false
  baseURL: "https://api.cluster2.orcunuso.io:6443"
  token: "eyJhbGciOiJSIk94.........RWNRb21fRHMwQkxBhdXliT2NQWF"
  quota: "quota-compute"
sheets:
  alerts: true
  nodes: true
  machines: true
  namespaces: true
  nsquotas: true
  services: true
  routes: true
  pvolumes: true
```

* The best practice is to create a resource quota resource within every namespace, preferably with the help of default project template. So "quota" key defines that default resource quota in that cluster.
* The booleans under "sheets" define if we need to get data from related resources. If true, OCinfo will create seperate sheets for every item.

### Creating service accounts

OCinfo needs a service account that has readonly permissions to get data from OpenShift and **cluster-reader** cluster role is a very proper candidate for this task but there is only one requirement that this role does not fulfill. OCinfo needs to read secrets, to be more specific, needs to get the token of openshift-monitoring/prometheus-k8s service account. So we need to create a custom cluster role that we can derive from cluster-reader and only grant additional read permissions to secrets.

This task needs to be performed on every OpenShift cluster that we will extract data from. 

```bash
oc apply -f conf/clusterrole-ocinfo.yaml
oc create project ocinfo
oc create sa ocinfo -n ocinfo
oc adm policy add-cluster-role-to-user custom-ocinfo -z ocinfo --rolebinding-name=custom-ocinfo -n ocinfo
```

Finally, we need to get the token of our service account and add it into our YAML file.

```bash
oc sa get-token ocinfo -n ocinfo
```

## Sample Execution

Once you have fulfilled the requirements, you can get the result spreadsheet document by just running the binary. Below is a screenshot on what the resulting file would look like:

![OCinfo Screenshot](https://github.com/vorcunus/ocinfo/blob/main/png/ocinfo-sshot1.png?raw=true)

### Gotchas

* In NSQuotas sheet, only the namespaces with quotas are listed in the sheet in order to make it possible to compare quota definitions with active usage. The units used for the cpu and memory metrics are millicores and mebibytes respectively. 

## Known issues and roadmap

This tool is still at early stages and there are many improvements on my mind from which I can list a few as;

* Increase the variety of resources to extract (like ingresses, deployments, daemonsets, etc)
* Summary sheet is left blank intentionally. Future releases will have pivots and charts to summarize sheets.
* OpenShift 4.6 support, especially with OVNKubernetes CNI.
* Query default Prometheus instance, scrape more data and provide metrics in proper sheets

## Contribution

Any contribution or suggestion that would help this tool more useful and efficient are welcome!

## Contact

Ozan Orçunus [@orcunuso](https://twitter.com/orcunuso)