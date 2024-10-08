# node-role-labeler

Operator that monitors nodes labels and sync custom role labels with official Kubernetes ones.

## Description

Kubernetes does not allow to assign `node-role.kubernetes.io/*` labels from kubelet, so Node cannot mark itself. Using these operator you can create `node-role.cluster.local/*` instead and operator will create respective `node-role.kubernetes.io/*` label and vice-versa. You can change this custom prefix using env variable `NODE_ROLE_PREFIX`, which is set to `node-role.cluster.local/` by default.

## Getting Started

Install using kustomize.

```shell
kubectl apply -f https://github.com/vlasov-y/node-role-labeler/releases/latest/download/install.yaml
```

## Demo

[![asciicast](https://asciinema.org/a/679600.svg)](https://asciinema.org/a/679600)

## License

[Apache License, Version 2.0](LICENSE)
