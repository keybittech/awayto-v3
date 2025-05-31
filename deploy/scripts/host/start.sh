#!/bin/bash

set -a
. etc-dir/.env 
set +a

exec /usr/local/bin/binary-name
