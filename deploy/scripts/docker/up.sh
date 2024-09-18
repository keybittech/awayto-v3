#!/bin/sh

cp ../../../.env .
. ./.env

docker volume create newpg15store
docker volume create newredisdata

docker compose up -d --build

sleep 10

sh ../auth/install.sh

rm ./.env
