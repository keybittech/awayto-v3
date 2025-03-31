#! /bins/h

# make doesn't rescan dependencies after the first function so this script splits them up

git pull
make build
make host_update
SUDO=sudo make docker_up
make host_deploy_op


