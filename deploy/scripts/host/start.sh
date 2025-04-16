#!/bin/bash

cd work-dir && make host_deploy_op && cd -

exec /usr/local/bin/binary-name
