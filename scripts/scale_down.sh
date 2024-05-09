#!/bin/sh

# Check if namespace argument is provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 <namespace>"
    exit 1
fi

namespace=$1

# Scale down deployments
deployments=$(kubectl get deployments -n $namespace -o jsonpath='{.items[*].metadata.name}')
for deployment in $deployments; do
    kubectl scale deployment $deployment --replicas=0 -n $namespace
    echo "Scaled down deployment $deployment"
done

# Scale down statefulsets
statefulsets=$(kubectl get statefulsets -n $namespace -o jsonpath='{.items[*].metadata.name}')
for statefulset in $statefulsets; do
    kubectl scale statefulset $statefulset --replicas=0 -n $namespace
    echo "Scaled down statefulset $statefulset"
done

echo "All deployments and statefulsets in namespace $namespace scaled down."

