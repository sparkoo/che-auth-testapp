#!/bin/bash

set -x

docker run --rm -p 8080:8080 -e KUBERNETES_SERVICE_HOST=$( minikube ip ) -e KUBERNETES_SERVICE_PORT=8443 -v /home/mvala/.minikube/ca.crt:/var/run/secrets/kubernetes.io/serviceaccount/ca.crt:z quay.io/mvala/che-auth-testapp:latest
