#!/bin/bash

cd work-dir && \
  SUDO=sudo make docker_up && \ 
  make host_deploy_op && \
  cd -

sleep 1
