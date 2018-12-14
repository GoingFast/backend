#!/bin/sh
set -eo pipefail

if [ -z "$APISERVER_ADDR" ]; then
    echo "Missing APISERVER_ADDR"
    exit 1
fi

if [ -z "$HELM_INSTALL" ]; then
    echo "Missing HELM_INSTALL"
    exit 1
fi

if [ -z "$CA" ]; then
    echo "Missing CA"
    exit 1
fi

if [ -z "$TOKEN" ]; then
    echo "Missing TOKEN"
    exit 1
fi

setup_kube() {
    echo "$TOKEN" | base64 -d > token
    echo "$CA" | base64 -d > ca.crt
    kubectl config set-credentials drone --token="$(cat ./token)"
    kubectl config set-cluster kubernetes --server="$APISERVER_ADDR" --certificate-authority=ca.crt
    kubectl config set-context drone --cluster=kubernetes --user=drone --namespace=default
    kubectl config use-context drone
    chmod 755 ~/.kube/config
    kubectl version
}

staging() {
    productionColor=$(helm get values --all "$HELM_INSTALL" | grep -Po "prod: \K.*" )
    stagingColor=$(helm get values --all "$HELM_INSTALL" | grep -Po 'staging: \K.*')
    printf "Staging: %s\\nProduction: %s" "$stagingColor\\n" "$productionColor"

    cd "$(git rev-parse --show-toplevel)/ci/charts/$HELM_INSTALL"
    helm upgrade --install $HELM_INSTALL . --reuse-values -f env.yaml --set "deployment.$stagingColor.enabled=true" --set "deployment.$stagingColor.client.image.tag=$TAG" --set "deployment.$stagingColor.server.image.tag=$TAG" "$ADDITIONAL_STAGING_ARGS"
}

switch() {
    productionColor=$(helm get values --all "$HELM_INSTALL" | grep -Po "prod: \K.*" )
    stagingColor=$(helm get values --all "$HELM_INSTALL" | grep -Po 'staging: \K.*')
    printf "Staging: %s\\nProduction: %s" "$stagingColor\\n" "$productionColor"

    cd "$(git rev-parse --show-toplevel)/ci/charts/$HELM_INSTALL"
    helm upgrade --install $HELM_INSTALL . --reuse-values -f env.yaml --set "deployment.prod=$stagingColor" --set deployment.staging="$productionColor" "$ADDITIONAL_SWITCH_ARGS"
}

setup_kube

if [ -n "$SWITCH" ]; then
    switch
fi

if [ -n "$STAGING" ]; then
    staging
fi
