#!/bin/sh

export RUN_ENV="prod"
RUN_NAME="pbh.btn.trunker"

#inject some environment variables

exec bin/${RUN_NAME}