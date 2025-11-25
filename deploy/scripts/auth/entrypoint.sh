#!/bin/bash

# From https://github.com/keycloak/keycloak/issues/10816

# Basic support for passing the db password as a mounted file. Or any other KC_ variable,
# really.
# Looks up environment variables like KC_*_FILE, reads the specified file and exports
# the content to KC_*
# e.g. KC_DB_PASSWORD_FILE -> KC_DB_PASSWORD

# give some time for the db entrypoint to finish
sleep 10

# Find suitable variables
lines=$(printenv | grep -o KC_.*_FILE)
# Split into array
vars=($lines)
# Enumerate variable names
for var in ${vars[@]}; do
    # Output variable, trim the _FILE suffix
    # e.g. KC_DB_PASSWORD_FILE -> KC_DB_PASSWORD
    outvar="${var%_FILE}"

    # Variable content = file path
    file="${!var}"

    # Empty value -> warn but don't fail
    if [[ -z $file ]]; then
        echo "WARN: $var specified but empty"
        continue
    fi

    # File exists
    if [[ -e $file ]]; then
        # Read contents
        content=$(cat $file)
        # Export contents if non-empty
        if [[ -n content ]]; then
            export $outvar=$content
            echo "INFO: exported $outvar from $var"
        # Empty contents, warn but don't fail
        else
            echo "WARN: $var -> $file is empty"
        fi
    # File is expected but not found. Very likely a misconfiguration, fail early
    else
        echo "ERR: $var -> file '$file' not found"
        exit 1
    fi    
done


# Pass all command parameters
exec /opt/keycloak/bin/kc.sh start --optimized "$@"
