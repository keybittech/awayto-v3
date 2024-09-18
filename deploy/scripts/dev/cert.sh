#!/bin/sh

if [ ! -f "certs/cert.pem" ]; then
  openssl req -nodes -new -x509 -keyout $KEY_LOC -out $CERT_LOC -days 365 -subj "/CN=$APP_HOST_NAME"
fi
