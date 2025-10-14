# Developer guide

A Kubernetes operator for managing financial resources and budget tracking.

## Description

Node role labeler is a Kubernetes operator that helps mirror custom roles labels to and from official `node-role.kubernetes.io/*` ones.

## Getting Started

### Prerequisites

- go version v1.24.0+
- docker version 17.03+
- [Task](https://taskfile.dev/) for build automation
- [Pre-commit](https://pre-commit.com/#install) for running git hooks on commit

Execute `pre-commit install` once before start of the development.

### To Deploy on the cluster

```sh
task kind:bootstrap install-operator
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall

This will uninstall the operator and CRDs if any, but leave the cluster running.

```sh
task uninstall-operator
```

To remove the cluster.

```sh
task kind:delete
```

>**NOTE**: When you run `task test:e2e` cluster will keep running as well. You should delete it manually if you want that.

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

This is usually done from CI pipelines.

```sh
task generate:distribute image=controller-image:with-tag
```

**NOTE:** The task commands mentioned above generate an 'install.yaml'.
This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

## Contributing

Open issues and PRs - we will review them and develop the application together!

**NOTE:** Run `task` for more information on all potential `task` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
