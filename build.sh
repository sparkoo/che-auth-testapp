#!/bin/bash

IMAGE=${IMAGE:-quay.io/mvala/che-auth-testapp:latest}

set -x

docker build -f build/Containerfile -t ${IMAGE} .
docker push ${IMAGE}
