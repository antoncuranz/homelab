#!/bin/sh

# Define variables
NAMESPACE="onepassword"
SECRET_NAME="onepassword-secret"

# Prompt user for credentials
read -rp "Enter the onepassword credentials (Kubernetes Credentials File, base64 enc.): " CREDENTIALS

# Prompt user for token
read -rp "Enter the onepassword token (Kubernetes Access Token): " TOKEN
echo # Move to a new line after token input

# Check if credentials and token are provided
if [ -z "$CREDENTIALS" ] || [ -z "$TOKEN" ]; then
    echo "Error: Missing one or more required inputs."
    exit 1
fi

# Create Kubernetes secret
kubectl create secret generic "$SECRET_NAME" \
    --namespace="$NAMESPACE" \
    --from-literal="credentials=$CREDENTIALS" \
    --from-literal="token=$TOKEN"

echo "Kubernetes secret '$SECRET_NAME' created in namespace '$NAMESPACE'."

