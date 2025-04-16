#!/bin/bash

set -a
. etc-dir/.env 
set +a
exec binary-name --log debug
