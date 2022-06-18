#!/usr/bin/env bash

echo "### Automatic approval ###"

ctx=$(kubectl config current-context)
echo "Current context is ${ctx}"

echo "Create temprolebinding"

kubectl apply -n test-namespace -f ./tests/automaticApproval/temprolebinding.yaml

sleep 3

phase=$(kubectl -n test-namespace get temprolebindings jonh-observe-duration -o json | jq -r ".status.phase")

echo "Phase is ${phase}"

if [ "$phase" == "Applied" ]; then
    echo "TRB applied"
else
    echo "TRB not applied , TEST FAILS"
    exit 1
fi

kubectl -n test-namespace get rolebinding jonh-observe-duration
if [ $? != 0 ]; then
    echo "Rolebinding not created, TEST FAILS"
    exit 1
fi

sleep 60

phase=$(kubectl -n test-namespace get temprolebindings jonh-observe-duration -o json | jq -r ".status.phase")

if [ "$phase" == "Expired" ]; then
    echo "TRB expired"
    kubectl -n test-namespace get rolebinding jonh-observe-duration
    if [ $? == 0 ];
    then
        echo "Rolebinding not deleted, TEST FAILS"
        exit 1
    fi
else
    echo "TRB not EXPIRED , TEST FAILS"
    exit 1
fi

echo "TEST SUCCESS"
exit 0