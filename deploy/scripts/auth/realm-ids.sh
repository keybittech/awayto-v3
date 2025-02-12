#!/bin/bash

# Input and output files
JSON_FILE="$(dirname "$0")/realm-complete.json.bak"

# Extract all unique IDs from the JSON file using jq
ids=$(jq -r '.. | .id? // empty' "$JSON_FILE" | sort | uniq)

# Process each ID
for old_id in $ids; do
    # Generate a new random UUID for the ID
    new_id=$(uuidgen | tr -d '-')

    echo "Changing ID $old_id to $new_id"

    # Replace the old ID with the new one in the JSON file
    jq --arg old_id "$old_id" --arg new_id "$new_id" \
        'walk(if type == "object" and .id == $old_id then .id = $new_id else . end)' \
        "$JSON_FILE" > tmp.json && mv tmp.json "$JSON_FILE"
    
    # Replace all references of the old_id with new_id throughout the file
    sed -i "s/$old_id/$new_id/g" "$JSON_FILE"
done

echo "IDs have been updated successfully."
