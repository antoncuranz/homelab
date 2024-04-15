#!/bin/sh

# Define variables
EXTERNAL_DNS_NAMESPACE="external-dns"
CERT_MANAGER_NAMESPACE="cert-manager"
SECRET_NAME="cloudflare-api-token"

# Prompt user for Cloudflare API token
echo "Create a Cloudflare API Token with All Zone.Zone:Read, Zone.DNS:Edit permissions."
echo "https://dash.cloudflare.com/profile/api-tokens"
read -rp "Enter Cloudflare API token: " CLOUDFLARE_API_TOKEN
echo

if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
    echo "Error: Missing Cloudflare API token."
    exit 1
fi

# Create Kubernetes secret for external-dns
kubectl create secret generic "$SECRET_NAME" \
    --namespace="$EXTERNAL_DNS_NAMESPACE" \
    --from-literal="value=$CLOUDFLARE_API_TOKEN"

echo "Kubernetes secret '$SECRET_NAME' created in namespace '$EXTERNAL_DNS_NAMESPACE' for external-dns."

# Create Kubernetes secret for cert-manager
kubectl create secret generic "$SECRET_NAME" \
    --namespace="$CERT_MANAGER_NAMESPACE" \
    --from-literal="api-token=$CLOUDFLARE_API_TOKEN"

echo "Kubernetes secret '$SECRET_NAME' created in namespace '$CERT_MANAGER_NAMESPACE' for cert-manager."

