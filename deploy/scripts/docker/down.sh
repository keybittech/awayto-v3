#!/bin/sh

docker compose down

docker volume remove newpg15store
docker volume remove newredisdata
