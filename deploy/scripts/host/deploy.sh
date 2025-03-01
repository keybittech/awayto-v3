#! /bins/h

# make doesn't rescan dependencies after the first function so this script splits them up

SUDO=sudo

make host_update
make build
make docker_up
make host_deploy_op


