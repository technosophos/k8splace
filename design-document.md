# Kubernetes Package Manager (KPM), Version 1

**TL;DR:** Homebrew for Kubernetes, with packages called "Kubes" (pronounced Cube)

This document outlines the architecture for a Kubernetes package management tool.

Our work suggests that great benefits could accrue if the community shares (and iterates on) a set of "best practices" Kubernetes packages. Examples of this might be:

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

In this document, we (tentatively) refer to a Kubernetes package as a Kube. The package format is discussed later in the document.

## A Package Archive

Right now, the Kubernetes community largely hand-crafts their manifest files. There is no accepted distribution schema, and the only commonly referenced set of files are the ones in the core Kubernetes GitHub repository.

An archive will provide a central place to store Kubes. Users will fetch information about available packages from this service.

## The Workflows

This section describes a few different workflows

### User Workflow (Simple)

Use case: User is looking for a standalone etcd instance (one pod, no clustering). User wants it immediately installed on Kubernetes.

```
$ kpm update
---> Fetching updates...
---> Done
$ kpm search etcd
	etcd-standalone: etcd-standalone is a single Etcd instance (no cluster)
	etcd-ha-cluster: an HA Etcd cluster (3+ pods)
	etcd-discovery:  a single-pod Etcd service discovery server
$ kpm info etcd-standalone
	Description: etcd-standalone is a single Etcd instance (no cluster)
	Version: 2.2.0-beta3
	Website: https://github.com/coreos/etcd
	Built: Oct. 21, 2015
	Provides: etcd-rc.yaml etcd-service.yaml
$ kpm install etcd-standalone
---> Downloading etcd-standalone-2.2.0-beta3
---> Cached files into $KPM_HOME/etcd-standalone-2.2.0-beta3/
---> Running kubectl create -f ...
---> Done
```

### User Workflow (Advanced)

Use Case: User wants to get an NGINX formula, modify it, and then load the modified version into a running Kubernetes cluster.

```
$ kpm update
---> Fetching updates...
---> Done
$ kpm get nginx-standard
---> Downloading nginx-standard
---> Cached files into $KPM_HOME/nginx-standard/
$ vi $KPM_HOME/nginx-standard/nginx-rc.yaml
$ kpm install nginx-standard
---> Running kubectl create -f ...
---> Done
```

TODO: If we back this to a strict Git workflow, we need a way to provide the user space for customized formulas.

## KPM Service Design

This section describes the design of a centralized service for KPM.

The goals for designing an archive backend are as follows:

- Supports authentication and authorization
- Provides simple methods for storing, retrieving, searching, and browsing
- Intuitive to a software developer
- Highly available
- Understandable API
- Provides workflow for submitting, reviewing, and filing bugs

### Option 1: Package Service Pulling From Many Git Repositories (NPM/Composer Model)

In this model, Deis manages a central repository service. The service is backed by a persistent database storage layer. This was the model for the original k8splace POC.

Individual packages are managed in their own git repositories, located at GitHub, Bitbucket, or any other Git service.

Contributors create KPM projects, and in so doing link the project to the package's Git repository. Using WebHooks or polling, the KPM server then synchronizes Git tags to the KPM server.

#### Pros

- Can fully automate the submission process (no need to manually review)
- Leverage Git
- Flexibility to add ratings system, comments, download count, etc.
- Issues filed against individual products, not against entire repo
- Contributors do not need to submit things directly to KPM

#### Cons

- Requires building or integrating an auth system
- Requires maintaining an HA server setup
- Must be coded from scratch
- Longer development time
- Must build a custom API
- Very little control over the quality of committed patches

### Option 2: GitHub Backend (Homebrew Model)

In this model, the only backend is a single GitHub repository. This repository contains all supported packages inside individual directories. This model was suggested after the POC project, and has a number of strong points that Option 1 does not have.

```
(root)
  |
  |- packages/
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

### Cons

- Features are limited to the features of GitHub. For example, no individual installation counts, download counts, etc.
- The repository itself must be monitored by "gatekeeper" community managers who will have to actively participate in reviewing and committing packages.
- Similar packages are hard to differentiate. For example, there may be multiple MySQL packages, each tuned 
- Requires user to proactively pull down the master repo

## A Command Line Client

The command line client is envisioned as the primary way that users interact with our system. It will be used for getting and installing packages.

The command line will support (at least) the following commands:


- `list`: List currently installed Kubes
- `search PATTERN`: Search the available Kubes, list those that match the `PATTERN`
- `get KUBE`: Retrieve the Kube and put it in a local working directory
	* For the homebrew model, this will copy the Kube resources to a place where the user can modify
	* For a central server model, this will download the package and install to a place where the user can modify.
- `install KUBE`: Get the Kube (if necessary) and then install it into Kubernetes, attaching to the k8s server configured for the current user via Kubectl.
- `info KUBE`: Print information about a Kube
- `update`: Refresh local metadata about available Kubes. On Homebrew model, this will re-pull the repo.
- `help`: Print help.

TODO: We would like to add a `test` command. How would this work?

## A Web Interface

The initial web interface will focus exclusively on education:

- Introduction to KPM
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

### The `K3.yaml` File

A Manifest file is a YAML-formatted file describing a package's purpose, version, and authority information. It is called `K3.yaml`, pronounced _k-cubed_ or _k-three_.

The manifest file contains the following information:

- name: The name of the package
- version: A SemVer 2 version number, no leading `v`
- home: The URL to a relevant project page, git repo, or contact person
- description: A single line description
- details: A multi-line description
- maintainers: A list of name and URL/email address combinations for the maintainer(s)
- dependencies: A list of other services on which this depends. Dependencies contain names and version ranges. See _Dependecies_ below.

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
```

#### Dependency Resolution

If dependencies are declared, `kpm` will...

- Check to see if the named dependency is locally present
- If it is, check to see if the version is within the supplied parameters

If either check fails, `kpm` will _emit a warning_, but will not prevent the installation. Under normal operation, it will prompt to continue. If the `--force` flag is set, it will simply continue.

Example:

```
$ kpm install ponycorn
!!-> WARNING: Dependency nginx (1.9 < x <= 1.10) does not seem to be satisfied.
!!-> This package may not function correctly.
Continue? yes
---> Running `kubectl create -f ...
```

### Manifests

All Kubernetes manifests are stored in a `manifests/` directory. They must be in either YAML or JSON, and
they must conform to the Kubernetes definitions.

Any pure YAML or JSON file must end with its appropriate suffix (`.yaml` or `.json`)

#### Manifest Templates

Packages may contain template files that the pre-install hooks can create.

- Templates MUST use Jinja-compatible template notation
- Templates MUST use the compound extension `.yaml.tmpl` or `.json.tmpl`
- Templates OUGHT NOT import other templates (please restrain yourself from creating template hierarchies)
- Authors OUGHT to provide a non-templated version whenever possible

Assuming, for example, that the `pre-install.py` passes in the sufficient template variables, the following template describes a YAML-formatted pod:

```jinja2
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
      imagePullPolicy: {{ mc.pullPolicy }}
      image: deis/mc:2015-sept
      env:
        {% for key, value in mc.env %}
        {{key}}: {{value}}
        {% endfor %}
```

### Pre-install Hooks

> This functionality will require the following constraints on a system:
>
> - Must have Python
> - Must have libraries for SSL support (openssl?)
> - Must have Jinja for text template support

A package maintainer may create a directory, called `scripts/`, which contains scripts to be executed on the local host (the system `kpm` is run on) prior to the manifests being pushed into Kubernetes. If this directory is present, `kpm` will execute any script named `pre-install.py`

The only supported scripting language for these is Python.

Pre-install hooks may be used to:

- generate keypairs, secrets, and other credentials
- collect user input for the purposes of modifying manifests
- render Jinja `tpl` files into Kubernetes JSON or YAML files

Pre-install hooks ought not...

- modify files outside the package
- download additional Python code

## Non-Features / Deferred Features

The following items have already been considered and will not be part of Version 1:

- Homebrew tap/keg system
- post-install exec scripts (run `kubectl exec -it` on a pod)

Will not do:

- Lua, Ruby, PHP as the scripting language: Will not do for maintenance and compatibility reasons
- strings.Template template library: No control structures, so will not do


