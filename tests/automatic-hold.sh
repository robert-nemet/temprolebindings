#!/usr/bin/env bash

echo "### Automatic approval ###"

ctx=$(kubectl config current-context)
echo "Current context is ${ctx}"

echo "Create temprolebinding"

kubectl apply -n test-namespace -f ./tests/automaticApproval/trb-hold.yaml

sleep 3

phase=$(kubectl -n test-namespace get temprolebindings jonh-observe-hold -o json | jq -r ".status.phase")

echo "Phase is ${phase}"

if [ "$phase" == "Hold" ]; then
    echo "TRB HOLD"
else
    echo "TRB not hold , TEST FAILS"
    exit 1
fi

echo "TEST SUCCESS"
exit 0