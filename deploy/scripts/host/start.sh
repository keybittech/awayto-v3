#!/bin/bash

set -a
. etc-dir/.env 
set +a
cd work-dir && SUDO=sudo make docker_up && cd -
cd work-dir && make host_service_start_op && cd -
