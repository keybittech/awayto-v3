#!/bin/sh
# --bind=$(grep APP_HOST .env | cut -d '=' -f2) 

# with ngrok
# sudo hugo serve --liveReloadPort 443 --port=443 --appendPort=false --baseURL=$(grep APP_HOST_URL .env | cut -d '=' -f2) -s . -d ./public

# with local
hugo serve
