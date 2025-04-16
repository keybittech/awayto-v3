#! /bins/h

# make doesn't rescan dependencies after the first function so this script splits them up

git reset --hard HEAD
git pull
make build
make host_update
make host_deploy_op

exit 0
