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

## Workflow

* create TRB
  * WH adds pending annotation
  * CTRL set status PENDING if annotation is PENDING

* Update TRB (annotation set to ACCEPT)
  * validate if prev state PENDING 
    * create RB (add labels/annotations to match TRB)
    * annotation and status set to APPLIED
    * queue new event

* Recheck (annotation set to APPLIED)
  * if expired 
    * delete RB (check labels/annotations to match TRB)
    * set annotation and status to EXPIRED
  * not expired: requeue

* Update TRB (annotation set to DENIED)
  * validate if status id PENDING
    * update status to DENIED
  * else do not allow change annotation

## History

* create project
    `kubebuilder init --domain rnemet.dev --repo rnemet.dev/temprolebindings --plugins go/v3 --project-version 3`
* init api: TempRoleBinding
    `kubebuilder create api --group tmprbac --version v1 --kind TempRoleBinding`
* init webhook:
    `kubebuilder create webhook --group tmprbac --version v1 --kind TempRoleBinding --defaulting --programmatic-validation`

