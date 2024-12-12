#!/usr/bin/env bash

export RUN_ENV="prod"
RUN_NAME="pbh.btn.server"

#inject some environment variables

exec bin/${RUN_NAME}