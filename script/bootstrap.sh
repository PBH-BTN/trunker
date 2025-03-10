#!/bin/sh

if [ -z "$RUN_ENV" ]; then
    export RUN_ENV="prod"
fi
RUN_NAME="pbh.btn.trunker"

#inject some environment variables

exec bin/${RUN_NAME}