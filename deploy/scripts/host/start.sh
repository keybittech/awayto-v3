#!/bin/bash

cd work-dir && SUDO=sudo make docker_up && make host_service_start_op && cd -

sleep 1

exit 0
