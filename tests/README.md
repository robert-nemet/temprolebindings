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

## Bugs

* [X] Deleting TempRoleBinding in Pending Status tries to delete RoleBindig