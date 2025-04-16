#!/bin/bash

SUDO=sudo cd work-dir && \
  make docker_up && \ 
  make host_service_start_op && \
  cd -
