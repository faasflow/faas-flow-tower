#!/bin/bash

# Check if docker is installed
if ! [ -x "$(command -v docker)" ]; then
  echo 'Unable to find docker command, please install Docker (https://www.docker.com/) and retry' >&2
  exit 1
fi

# Check if faas-cli is installed
if ! [ -x "$(command -v faas-cli)" ]; then
  echo 'Unable to find faas command, please install faas-cli (https://docs.openfaas.com/cli/install/) and retry' >&2
  exit 1
fi

echo "building faas function"
faas-cli build
echo "pulling images"
docker pull s8sg/consul
docker pull minio/minio:latest
docker pull jaegertracing/all-in-one:1.8
