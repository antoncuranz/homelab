#!/bin/sh

NAMESPACE="onepassword"
SECRET_NAME="onepassword-secret"

read -rp "Enter path to 1Password credentials file: " CREDENTIALS_FILE

read -rp "Enter the onepassword token (Kubernetes Access Token): " TOKEN
echo # Move to a new line after token input

if [ -z "$CREDENTIALS_FILE" ] || [ -z "$TOKEN" ]; then
    echo "Error: Missing one or more required inputs."
    exit 1
fi

if [ ! -f "$CREDENTIALS_FILE" ]; then
    echo "Error: Credentials file not found: $CREDENTIALS_FILE"
    exit 1
fi

kubectl create secret generic "$SECRET_NAME" \
    --namespace="$NAMESPACE" \
    --from-file="credentials=$CREDENTIALS_FILE" \
    --from-literal="token=$TOKEN" \
    --dry-run=client -o yaml | kubectl apply -f -

echo "Kubernetes secret '$SECRET_NAME' applied in namespace '$NAMESPACE'."
