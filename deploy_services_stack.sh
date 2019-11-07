#!/bin/sh

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

# Create network func_function if doesn't exists
echo "Creating Function Network (func_functions) if doesn't exist"
[ ! "$(docker network ls | grep func_functions)" ] && docker network create -d overlay --attachable --label "openfaas=true" func_functions

# Secrets should be created for minio access.
echo "Attempting to create credentials for minio.."
SECRET_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
ACCESS_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
echo -n "$SECRET_KEY" | docker secret create s3-secret-key -
echo -n "$ACCESS_KEY" | docker secret create s3-access-key -
if [ $? = 0 ];
then
    echo "[Minio Credentials]\n Secret Key: $SECRET_KEY \n Access Key: $ACCESS_KEY"
else
    echo "[Minio Credentials]\n already exist, not creating"
fi

echo "Deploying faas-flow stack"
docker stack deploy --compose-file docker-compose.yml faasflow
faas-cli deploy
