#!/usr/bin/env bash

echo "#### Setup tests ####"

ctx=$(kubectl config current-context)

echo "Current context is ${ctx}"
echo "Current working directory is $(pwd)"

echo "Installing requiremets"

kubectl apply -f ./tests/setup/requirements.yaml
kubectl certificate approve john
kubectl get ns test-namespace
