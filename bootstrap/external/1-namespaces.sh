#!/bin/sh
kubectl create namespace cert-manager --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace external-dns --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace onepassword --dry-run=client -o yaml | kubectl apply -f -
