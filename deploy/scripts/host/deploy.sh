#! /bins/h

# make doesn't rescan dependencies after the first function so this script splits them up

SUDO=sudo

git pull
make build
make host_update
make docker_up
make host_deploy_op


