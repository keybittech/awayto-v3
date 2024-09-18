#!/bin/sh

if ! command -v docker >/dev/null 2>&1; then
  sudo mkdir -p /etc/docker
  curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh
  sudo systemctl restart docker
fi
