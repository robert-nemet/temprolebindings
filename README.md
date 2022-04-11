# Temprolebindings

## Motivation

During daily work, there is a need to grant temporary access to developers and/or other operators. TempRoleOperator control
a process of creating, approving, declining, and expiring RoleBinding. 

## Process

### create cluster

`k3d cluster create --config cluster/k3d.yml`

After modifying API: `make manifests`

Install: `make install`

Run: `make run`

## Deploy 

Install cert-manager first. Be sure that `imagePullPolicy` is `IfNotPresent`.

`make deploy IMG=rnemet/simpleapp`

`make docker-build IMG=rnemet/simpleapp`

`k3d image import -c playground rnemet/simpleapp`

## Cert manager

`kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml`



## History

* create project
    `kubebuilder init --domain rnemet.dev --repo rnemet.dev/temprolebindings --plugins go/v3 --project-version 3`
* init api:
  * TempRoleBinding `kubebuilder create api --group tmprbac --version v1 --kind TempRoleBinding`
  * TempClusterRoleBinding `kubebuilder create api --group tmprbac --version v1 --kind TempClusterRoleBinding`
* init webhook:
    `kubebuilder create webhook --group tmprbac --version v1 --kind TempRoleBinding --defaulting --programmatic-validation`

