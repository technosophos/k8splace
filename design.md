# Helm: The Kubernetes Package Manager, Version 1

**TL;DR:** Homebrew for Kubernetes, with packages called "charts"

This document outlines the architecture for a Kubernetes package management tool. Drawing on Kubernetes' nautical heritage, the project is named _Helm_, and the command line client is named `helm`.

Our work suggests that great benefits could accrue if the community shares (and iterates on) a best practice for Kubernetes packages. Examples of this might be:

- A Postgres database
- A cluster of Memcached servers
- A Varnish caching proxy

By providing common service definitions, lifecycle handling, health checking, and configuration, a Kubernetes package can serve the purposes of many Kubernetes users, and serve as a foundation for many other users.

## What Is a Kubernetes Package?

For our purposes, an "application" is any pre-packaged set of manifest files that can be deployed into Kubernetes, and that implement a particular functional unit of service (e.g. a web application, a database, or a message queue).

When it comes to deploying an application into Kubernetes there are always one or more Kubernetes manifest files. Often, these files (particularly pod and replication controller manifests) will reference one or more container images. These are typically pulled from external registries like GCR, Quay.io, and Docker Hub.

A Kubernetes package is thus:

- A collection of one or more  Kubernetes manifest files...
- That can be installed together...
- Together with descriptive metadata...
- And possibly some configuration scripts

In this document, we refer to a Helm Kubernetes package as a *Chart*. The package format is discussed later in the document.

## A Package Archive

Right now, the Kubernetes community largely hand-crafts their manifest files. There is no accepted distribution schema, and the only commonly referenced set of files are the ones in the core Kubernetes GitHub repository.

An archive will provide a central place to store Charts. Users will fetch information about available packages from this service.

## The Workflows

This section describes a few different workflows. Two workflows are presented: On describes the most basic workflow, while a second describes what we think is the more common workflow.

### User Workflow (Simple)

Use case: User is looking for a standalone etcd instance (one pod, no clustering). User wants it immediately installed on Kubernetes.

Assuming a user has already configured the environment with `kubectl`, the workflow will be something like this:

```
$ helm update
---> Fetching updates...
---> Done
$ helm search etcd
	etcd-standalone: etcd-standalone is a single Etcd instance (no cluster)
	etcd-ha-cluster: an HA Etcd cluster (3+ pods)
	etcd-discovery:  a single-pod Etcd service discovery server
$ helm info etcd-standalone
	Description: etcd-standalone is a single Etcd instance (no cluster)
	Version: 2.2.0-beta3
	Website: https://github.com/coreos/etcd
	Built: Oct. 21, 2015
	Provides: etcd-rc.yaml etcd-service.yaml
$ helm install etcd-standalone
---> Downloading etcd-standalone-2.2.0-beta3
---> Cached files into $helm_HOME/etcd-standalone-2.2.0-beta3/
---> Running kubectl create -f ...
---> Done
```

### User Workflow (Advanced)

Use Case: User wants to get an NGINX formula, modify it, and then load the modified version into a running Kubernetes cluster.

```
$ helm status
	Kubernetes API Server: https://10.21.77.3:443
	Namespace: technosophos
	User: mbutcher
	Cluster: demo
$ helm update
---> Fetching updates...
---> Done
$ helm fetch nginx-standard
---> Downloading nginx-standard
---> Cached files into $helm_HOME/nginx-standard/
$ vi $HELM_HOME/workspace/nginx-standard/manifests/nginx-rc.yaml
$ helm install nginx-standard
---> Running kubectl create -f ...
---> Done
```

## helm Service Design

This section describes the design of a centralized service for helm.

The goals for designing an archive backend are as follows:

- Supports authentication and authorization
- Provides simple methods for storing, retrieving, searching, and browsing
- Intuitive to a software developer
- Highly available
- Understandable API
- Provides workflow for submitting, reviewing, and filing bugs

The service will use GitHub directly as a single repository backend, much like [homebrew](http://brew.sh) does.

In this model, the repository contains all supported packages inside individual directories. By leveraging GitHub's architecture and services, we can focus on managing the packages and submissions, rather than building and maintaining infrastructure.

```
(root)
  |
  |- charts/
       |
       |- mypackage/
       |
       |- postgres/
       |
       |- ...
```

The API to this service is essentially the git protocol.

Contributing packages to the archive is done by submitting Git pull requests.

### Pros

- Fast to set up
- No server maintenance
- Authentication and authorization provided automatically
- Highly available already
- Workflow is well understood
- Strong tooling, no new API
- Easy integration with single CI/CD system
- The repository itself is monitored by "gatekeeper" community managers who will have to actively participate in reviewing and committing packages.


### Cons

- Features are limited to the features of GitHub. For example, no individual installation counts, download counts, etc. _Strategy:_ Investigate building an API server to provide metrics data.
- Similar packages are hard to differentiate. For example, there may be multiple MySQL packages, each tuned For a specific use case. _Strategy:_ Lay out clear guidelines on package naming.
- Requires user to proactively pull down the master repo. _Strategy:_ The tool warns users when they have not updated in a long time.

## A Command Line Client

The command line client is envisioned as the primary way that users interact with our system. It will be used for getting and installing packages.

The command line will support (at least) the following commands:


- `list`: List currently installed Charts
- `search PATTERN`: Search the available Charts, list those that match the `PATTERN`
- `fetch CHART`: Retrieve the Kube and put it in a working directory. The working directory is where pre-install scripts are run, and where users may modify by hand.
- `install CHART`: Get the Kube (if necessary) and then install it into Kubernetes, attaching to the k8s server configured for the current user via Kubectl. This reads from the working directory.
- `info CHART`: Print information about a Kube.
- `update`: Refresh local metadata about available Charts. On Homebrew model, this will re-pull the repo.
- `verify`: Deploy the package into a cluster, and then validate that it completes its installation process. This is intended for chart developers.
- `help`: Print help.

The client will locally track the following three things:

- The client's own configuration
- The packages available from the remote service (a local cache)
- The packages that the client has "downloaded"

The tree looks something like this:

```
- $HELM_HOME
      |
      |- cache/       # Where `helm update` data goes. The content of this directory
      |               # is a git repo.
      |
      |- workdir/     # Working directory, where `helm get` copies go.
      |
      |- config.yaml  # configuration info
```

The default location for `$HELM_HOME` will be `$HOME/.helm`. This reflects the fact that package management by this method is scoped to the developer, not to the system. However, we will make this flexible because we have the following target use cases in mind:

- Individual helm developer/user
- CI/CD system
- Workdir shared by `git` repo among several developers
- Dockerized helm that runs in a container

### Client Config

Clients will track minimal configuration about remote hosts, local state, and possibly security.

The choice of services will determine what directives are needed in this file (if any).

## A Web Interface

The initial web interface will focus exclusively on education:

- Introduction to helm
- Instructions for installing
- Instructions for package submission

In the future, we will support the following additional features:

- Package search: Search for packages and show the details
- Test results: View the test results for any package

## CI/CD Service

To maintain a high bar for the submitted projects, a CI/CD service is run on each submitted project. The initial CI/CD implementation does the following:

- Runs the install script on an isolated container, and ensures that it exits with no errors
- Validates the manifest files
- Runs the package on its own in a Kubernetes cluster
- Verifies that the package can successfully enter the "Running" state (e.g. passes initial health checks)

## The Package Format

A package will be composed of the following parts:

- A metadata file
- One or more Kubernetes manifest files, stored in a `manifests/` directory
- Zero or more pre-install scripts (executed on the local host prior to pushing to Kubernetes)

### The `Chart.yaml` File

A Manifest file is a YAML-formatted file describing a package's purpose, version, and authority information. It is called `Chart.yaml`.

The manifest file contains the following information:

- name: The name of the package
- version: A SemVer 2 version number, no leading `v`
- home: The URL to a relevant project page, git repo, or contact person
- description: A single line description
- details: A multi-line description
- maintainers: A list of name and URL/email address combinations for the maintainer(s)
- dependencies: A list of other services on which this depends. Dependencies contain names and version ranges. See _Dependecies_ below.
- A pre-install command chain (like Ansible playbooks) to generate any necessary template variables

Example:

```yaml
name: ponycorn
home: http://github.com/technosophos/ponycorn
version: 1.0.3
description: The ponycorn server
maintainers:
  - Matt Butcher <mbutcher@deis.com>
dependencies:
  - name: nginx
    version: 1.9 < x <= 1.10
details:
  This package provides the ponycorn server, which requires an existing nginx
  server to function correctly.
pre-install:
  - generate-keypair foo
```

#### Dependency Resolution

If dependencies are declared, `helm` will...

- Check to see if the named dependency is locally present
- If it is, check to see if the version is within the supplied parameters

If either check fails, `helm` will _emit a warning_, but will not prevent the installation. Under normal operation, it will prompt to continue. If the `--force` flag is set, it will simply continue.

Example:

```
$ helm install ponycorn
!!-> WARNING: Dependency nginx (1.9 < x <= 1.10) does not seem to be satisfied.
!!-> This package may not function correctly.
Continue? yes
---> Running `kubectl create -f ...
```

### Manifest Templates

All Kubernetes manifests templates are stored in a `templates/` directory. They must be in either YAML or JSON (plus template directives), and they must conform to the Kubernetes definitions.

- Templates MUST use Go `text/template` format
- Templates MUST use the compound extension `.yaml.tmpl` or `.json.tmpl`

Assuming, for example, that the `pre-install` step passes in the sufficient template context, the following template describes a YAML-formatted pod:

```
apiVersion: v1
kind: Pod
metadata:
  name: deis-mc
  labels:
    heritage: deis
    version: 2015-sept
spec:
  restartPolicy: Never
  containers:
    - name: mc
      imagePullPolicy: {{ .PullPolicy }}
      image: deis/mc:2015-sept
      env:
        {{ range $key, $value := .Env }}
        {{$key}}: {{$value}}
        {{ end }}
```

### Pre-install Hooks

The Helm system will provide a number of built in tools for generating variables that will later be injected into templates. During the pre-install phase, Helm will read in a configuration YAML file, execute any necessary scripts, and then inject the data into the template's context.

Pre-install hooks may be used to:

- generate keypairs, secrets, and other credentials
- collect user input for the purposes of modifying manifests
- render `tpl` files into Kubernetes JSON or YAML files

## Non-Features / Deferred Features

The following items have already been considered and will not be part of Version 1:

- Homebrew tap/keg system
- post-install exec scripts (run `kubectl exec -it` on a pod)

Will not do:

- Lua, Ruby, PHP as the scripting language: Will not do for maintenance and compatibility reasons
- strings.Template template library: No control structures, so will not do


