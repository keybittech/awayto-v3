#!/bin/sh

if [ ! -f "$CERT_LOC" ]; then
  openssl req -nodes -new -x509 -keyout "$KEY_LOC" -out "$CERT_LOC" -days 365 -subj "/CN=$APP_HOST_NAME"
  echo "pwd required to update cert chain"
  sudo cp "$CERT_LOC" /usr/local/share/ca-certificates
  sudo update-ca-certificates
fi
