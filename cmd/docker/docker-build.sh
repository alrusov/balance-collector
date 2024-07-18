#!/bin/bash

set -e

#----------------------------------------------------------------#

if [ -f .env ]; then
	. .env
fi

#----------------------------------------------------------------#

BUILD_HOME=___DOCKER_TMP___

APP=$(basename $(readlink -f ../../))
APP_HOME=/opt/${APP}

TAG_LATEST="${DOCKER_REGISTRY}/${APP}:${VERSION:=latest}"


rm -rf ${BUILD_HOME}
mkdir -p ${BUILD_HOME}/
cp -r ../../balance-collector ../../config ../../html ../../templates ${BUILD_HOME}/
rm -f ${BUILD_HOME}/config/entities.toml
touch entities.toml balance-collector.db

HEALTHCHECK=". .env && curl -fs http://localhost:${PORT}/ || exit 1"

docker build \
	-t ${TAG_LATEST} \
	--build-arg BUILD_HOME=${BUILD_HOME} \
	--build-arg APP_HOME=${APP_HOME} \
	--build-arg APP=${APP} \
	--build-arg DOCKER_USER=${DOCKER_USER} \
	--build-arg DOCKER_UID=${DOCKER_UID} \
	--build-arg DOCKER_GROUP=${DOCKER_GROUP} \
	--build-arg DOCKER_GID=${DOCKER_GID} \
	--build-arg DOCKER_TZ=${DOCKER_TZ} \
	--build-arg HEALTHCHECK="${HEALTHCHECK}" \
	.

rm -rf ${BUILD_HOME}

#----------------------------------------------------------------#

#if [ "$1" == "push" -o "$1" == "" ]; then
#	docker push ${TAG_LATEST}
#fi

#----------------------------------------------------------------#
