#!/bin/sh

export RUN_ENV="prod"
RUN_NAME="pbh.btn.trunker"

#inject some environment variables

exec bin/${RUN_NAME} &
pid=$!
# Trap the SIGTERM signal and forward it to the main process
trap 'kill -SIGINT $pid; wait $pid' SIGTERM

# Wait for the main process to complete
wait $pidh