#!/bin/sh

AUTH_CID=$($SUDO docker ps -aqf "name=auth")
kcadm() {
  $SUDO docker exec $AUTH_CID /bin/sh opt/keycloak/bin/kcadm.sh "$@"
}

echo "# Waiting for keycloak to start."

kc_connected=1

while [ $kc_connected -eq 1 ]; do
  kcadm config credentials --server $KC_INTERNAL --realm master --user $KC_ADMIN --password $(cat $KC_PASS_FILE) 2> /dev/null
  kc_connected=$?
  sleep 2 # sleep for 2 seconds before the next attempt
done

kcadm update realms/master -s sslRequired=NONE

EXISTING=$(kcadm get realms/$KC_REALM)

if [ "$EXISTING" != "" ]; then
  echo "Exiting due to existing realm"
  exit 0
fi

# Realm creation
echo "# Creating the realm $KC_REALM."
kcadm create realms -s realm=$KC_REALM -s enabled=true
kcadm update realms/$KC_REALM \
  -s registrationAllowed=true \
  -s resetPasswordAllowed=true \
  -s rememberMe=true \
  -s registrationEmailAsUsername=true \
  -s loginWithEmailAllowed=true \
  -s loginTheme=site \
  -s adminTheme=kcv2 \
  -s emailTheme=site

echo "# Configuring authenticator to use custom registration."
kcadm create authentication/flows/registration/copy -r $KC_REALM -s newName="custom registration"

REGISTRATION_USER_CREATION_ID=$(kcadm get authentication/flows/custom%20registration/executions -r $KC_REALM | jq -r '.[] | select(.displayName == "Registration User Profile Creation") | .id')

kcadm delete authentication/executions/$REGISTRATION_USER_CREATION_ID -r $KC_REALM

CUSTOM_REGISTRATION_USER_CREATION_ID=$(kcadm create authentication/flows/custom%20registration%20registration%20form/executions/execution -r $KC_REALM -s provider=custom-registration-user-creation -i)

kcadm create authentication/executions/$CUSTOM_REGISTRATION_USER_CREATION_ID/raise-priority -r $KC_REALM
kcadm create authentication/executions/$CUSTOM_REGISTRATION_USER_CREATION_ID/raise-priority -r $KC_REALM
kcadm create authentication/executions/$CUSTOM_REGISTRATION_USER_CREATION_ID/raise-priority -r $KC_REALM

kcadm update authentication/flows/custom%20registration/executions -r $KC_REALM -b '{"id":"'"$CUSTOM_REGISTRATION_USER_CREATION_ID"'","requirement":"REQUIRED"}'

kcadm update realms/$KC_REALM -b '{ "registrationFlow": "custom registration", "attributes": { "userProfileEnabled": true } }'

echo "# Configuring front-end auth client."
SITE_CLIENT_ID=$(kcadm create clients -r $KC_REALM -s clientId=$KC_CLIENT -s 'redirectUris=["'"$APP_HOST_URL/*"'"]' -s rootUrl=$APP_HOST_URL -s baseUrl=$APP_HOST_URL -s publicClient=true -s standardFlowEnabled=true -s directAccessGrantsEnabled=true -s attributes='{ "post.logout.redirect.uris": "'"$APP_HOST_URL"'", "access.token.lifespan": 60 }' -i)
echo "Site Client ID "$SITE_CLIENT_ID

echo "# Attaching roles."
# GROUP_FEATURES
SITE_ROLES="GROUP_ADMIN GROUP_BOOKINGS GROUP_PERMISSIONS GROUP_ROLES GROUP_SCHEDULES GROUP_SCHEDULE_KEYS GROUP_SERVICES GROUP_USERS ROLE_CALL"
for SITE_ROLE in $SITE_ROLES; do
  kcadm create clients/$SITE_CLIENT_ID/roles -r $KC_REALM -s name=APP_$SITE_ROLE
done

echo "# Create group scope element with attribute mapper."
GROUP_SCOPE_ID=$(kcadm create client-scopes -r $KC_REALM -b '{"name":"groups","description":"","attributes":{"consent.screen.text":"","display.on.consent.screen":"true","include.in.token.scope":"true","gui.order":""},"type":"none","protocol":"openid-connect"}' -i)

kcadm update default-default-client-scopes/$GROUP_SCOPE_ID -r $KC_REALM

kcadm create client-scopes/$GROUP_SCOPE_ID/protocol-mappers/models -r $KC_REALM -b '{"protocol":"openid-connect","protocolMapper":"oidc-group-membership-mapper","name":"groups","config":{"claim.name":"groups","full.path":"true","id.token.claim":false,"access.token.claim":"true","userinfo.token.claim":false}}'

kcadm update clients/$SITE_CLIENT_ID/default-client-scopes/$GROUP_SCOPE_ID -r $KC_REALM

echo "# Configuring api client."
API_CLIENT_ID=$(kcadm create clients -r $KC_REALM -s clientId=$KC_API_CLIENT -s standardFlowEnabled=true -s serviceAccountsEnabled=true -i)

kcadm add-roles -r $KC_REALM --uusername service-account-$KC_API_CLIENT --cclientid realm-management --rolename manage-clients --rolename manage-realm --rolename manage-users

kcadm update clients/$API_CLIENT_ID -r $KC_REALM -s "secret=$(cat $KC_API_CLIENT_SECRET_FILE)"

echo "# Keycloak configuration finished."

exit 0
