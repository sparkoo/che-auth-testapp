#!/bin/bash
set -x

mkdir -p ~/.minikube/files/etc/ca-certificates/
cp minikube_users.csv ~/.minikube/files/etc/ca-certificates/tokens.csv

minikube --extra-config="apiserver.token-auth-file=/etc/ca-certificates/tokens.csv" start
minikube addons enable ingress
