#!/bin/bash

set -a
. etc-dir/.env 
set +a

cd work-dir && SUDO=sudo make docker_up && cd -

exec /usr/local/bin/binary-name
