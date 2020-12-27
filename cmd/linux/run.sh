#!/bin/bash

source $(dirname $0)/env.sh
${APP_EXEC} --config ../../config/${APP}.toml
