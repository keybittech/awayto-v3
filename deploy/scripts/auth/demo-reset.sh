#!/bin/sh

AUTH_CID=$($SUDO docker ps -aqf "name=auth")
kcadm() {
  $SUDO docker exec $AUTH_CID /bin/sh opt/keycloak/bin/kcadm.sh "$@"
}

kc_connected=0;

while [ $kc_connected -eq 1 ]; do
  kcadm config credentials --server $KC_INTERNAL --realm master --user $KC_ADMIN --password $(cat $KC_PASS_FILE)
  kc_connected=$?
  sleep 2
done

GROUP_ID=$(kcadm get groups?search=the_test_group -r $KC_REALM | jq -r '.[] | .id')
kcadm delete groups/$GROUP_ID -r $KC_REALM
echo "Deleted demo group"

for USER_ID in $(kcadm get users -r $KC_REALM | jq -r '.[] | .id' | xargs echo); do
  echo "Deleting demo user $USER_ID"
  kcadm delete users/$USER_ID -r $KC_REALM
done
