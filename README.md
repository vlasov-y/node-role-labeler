# Node Role Labeler

Operator that monitors nodes labels and sync custom role labels with official Kubernetes ones.

## Description

Kubernetes does not allow kubelet to assign `node-role.kubernetes.io/*` labels, so nodes cannot self-identify. With this operator, you can create `node-role.cluster.local/*` labels and the operator will automatically sync them with corresponding `node-role.kubernetes.io/*` labels and vice versa. You can change the default custom prefix using the environment variable `NODE_ROLE_PREFIX`, which is set to `node-role.cluster.local/` by default.

## Getting Started

Install using kustomize.

```shell
kubectl apply -f https://github.com/vlasov-y/node-role-labeler/releases/latest/download/install.yaml
```

## CLI args

There is a list of CLI args you can append to manager args in the deployment to tune the behaviour.

```shell
--burst int
  Burst to use while talking with kubernetes apiserver (default 30)
--enable-http2
  If set, HTTP/2 will be enabled for the metrics and webhook servers
--health-probe-bind-address string
  The address the probe endpoint binds to. (default ":8081")
--kubeconfig string
  Paths to a kubeconfig. Only required if out-of-cluster.
--leader-elect
  Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
--max-concurrent-reconciles int
  Maximum number of concurrent reconciles per reconciler (default 10)
--metrics-bind-address string
  The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service. (default "0")
--metrics-cert-key string
  The name of the metrics server key file. (default "tls.key")
--metrics-cert-name string
  The name of the metrics server certificate file. (default "tls.crt")
--metrics-cert-path string
  The directory that contains the metrics server certificate.
--metrics-secure
  If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead. (default true)
--qps float
  QPS to use while talking with kubernetes apiserver (default 20)
--webhook-cert-key string
  The name of the webhook key file. (default "tls.key")
--webhook-cert-name string
  The name of the webhook certificate file. (default "tls.crt")
--webhook-cert-path string
  The directory that contains the webhook certificate.
--zap-devel
  Development Mode defaults(encoder=consoleEncoder,logLevel=Debug,stackTraceLevel=Warn). Production Mode defaults(encoder=jsonEncoder,logLevel=Info,stackTraceLevel=Error) (default true)
--zap-encoder value
  Zap log encoding (one of 'json' or 'console')
--zap-log-level value
  Zap Level to configure the verbosity of logging. Can be one of 'debug', 'info', 'error', 'panic'or any integer value > 0 which corresponds to custom debug levels of increasing verbosity
--zap-stacktrace-level value
  Zap Level at and above which stacktraces are captured (one of 'info', 'error', 'panic').
--zap-time-encoding value
  Zap time encoding (one of 'epoch', 'millis', 'nano', 'iso8601', 'rfc3339' or 'rfc3339nano'). Defaults to 'epoch'.
```

Example kustomization to add arguments.

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - https://github.com/vlasov-y/node-role-labeler/releases/latest/download/install.yaml

## Uncomment if you want to disable creation of custom namespace
# namespace: kube-system
# patches:
#   - patch: |-
#       $patch: delete
#       apiVersion: v1
#       kind: Namespace
#       metadata:
#         name: _
#     target:
#       kind: Namespace

```

## Annotations

These annotation are set on nodes.

| Name                          | Default | Description                                                       |
| ----------------------------- | ------- | ----------------------------------------------------------------- |
| `node-role-labeler.io/state`  | *JSON*  | Controlled by operator, do not change                             |
| `node-role-labeler.io/enable` | *true*  | Set to "false" in order to disable operator for a particular node |

## Demo

[![asciicast](https://asciinema.org/a/679600.svg)](https://asciinema.org/a/679600)

## License

[Apache License, Version 2.0](LICENSE)
