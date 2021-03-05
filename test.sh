#!/bin/bash

URL=${1:-${URL}}
NAMESPACE=${2:-${NAMESPACE}}
TOKEN=${3:-${TOKEN}}

if [ -z "$URL" -o -z "$NAMESPACE"  -o -z "$TOKEN" ]; then
  echo
  echo "You must provide 3 parameters:"
  echo "$ ./test.sh <URL> <NAMESPACE> <TOKEN>"
  echo
  exit 1
fi

curl -H "Authorization: Bearer ${TOKEN}" ${URL}/${NAMESPACE}
