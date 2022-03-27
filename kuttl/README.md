# How to run tests

1. Start cluster (`k3d cluster create --config cluster/k3d.yml`)
2. Install cert-manager ( `kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml` )
3. Install operator (`make deploy IMG=rnemet/simpleapp`)
4. Change controller imagePullPolicy for controller image to IfNotPresent
5. Build docker image (`make docker-build IMG=rnemet/simpleapp`)
6. Push docker image to local cluster (`k3d image import -c playground rnemet/simpleapp`)
7. Restart controller deployment( ` k -n temprolebinding-system rollout restart deployment/temprolebinding-controller-manager `)
8. Run tests

## Test with kind

1. Start cluster: ` kind cluster create `
2. Install cert-manager ( `kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml` )
3. Install operator (`make deploy IMG=rnemet/simpleapp`)
4. Change controller imagePullPolicy for controller image to IfNotPresent
5. Build docker image (`make docker-build IMG=rnemet/simpleapp`)
6. Push docker image to local cluster (`kind load docker-image rnemet/simpleapp`)
7. Restart controller deployment( ` k -n temprolebinding-system rollout restart deployment/temprolebinding-controller-manager `)
8. Run tests

## Requirements

* [krew](https://krew.sigs.k8s.io/docs/user-guide/)
* [kuttl](https://kuttl.dev/)

## Webhook certificates

```shell
openssl genrsa -out tls.key 2048
openssl req -x509 -new -nodes -key tls.key -subj "/CN=172.18.0.2" -out tls.crt
```

# Test Scenario

| Name | Command(s) | Explanations |
|---|---|---|
| Create Pending | `k apply -f tmprbac_v1_temprolebinding.yaml` |Status should show all conditions. Only Pending should be true, and have set transitionTime. Phase should be set to Pending |
|Decline | `kubectl -n default annotate temprolebindings.tmprbac.rnemet.dev temprolebinding-sample  tmprbac/status=Declined --overwrite` | This sets annotation to Declined. Status condition Declined is true. Transition time is set. Phase is Declined |
| Approved | `kubectl -n default annotate temprolebindings.tmprbac.rnemet.dev temprolebinding-sample  tmprbac/status=Approved --overwrite` | This sets annotation to Approved. Status condition changed to Approved. Transition time set. Phase Approved |

