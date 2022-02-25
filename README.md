# Temprolebindings

## Motivation

TODO:

## Process

After modifying API:
    `make manifests`

Install: `make install`

Run: `make run`

## Deploy 

`make deploy IMG=rnemet/simpleapp`

`make docker-build IMG=rnemet/simpleapp`

`k3d image import -c playground rnemet/simpleapp`

Be sure pullImagePolicy is IfNotPresent.

## Cert manager

`kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml`

## History

* create project
    `kubebuilder init --domain rnemet.dev --repo rnemet.dev/temprolebindings --plugins go/v3 --project-version 3`
* init api: TempRoleBinding
    `kubebuilder create api --group tmprbac --version v1 --kind TempRoleBinding`
* init webhook:
    `kubebuilder create webhook --group tmprbac --version v1 --kind TempRoleBinding --defaulting --programmatic-validation`