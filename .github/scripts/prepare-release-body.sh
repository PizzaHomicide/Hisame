#!/bin/bash

VERSION="$1"
TEMPLATE_PATH="$2"
CHANGELOG_PATH="$3"
OUTPUT_PATH="$4"

# Extract changelog for this version
CHANGELOG=$(awk -v ver="${VERSION#v}" '
    $0 ~ "^## \\[" ver "\\]" {flag=1; next}
    $0 ~ "^## \\[" && $0 !~ ver {flag=0}
    flag {print}
' "$CHANGELOG_PATH")

# Read the template
TEMPLATE=$(cat "$TEMPLATE_PATH")

# Escape special characters for sed and prepare the changelog
ESCAPED_CHANGELOG=$(echo "$CHANGELOG" | sed -e 's/[\/&]/\\&/g' | sed -e ':a;N;$!ba;s/\n/\\n/g')

# Replace placeholders
RELEASE_BODY=$(echo "$TEMPLATE" |
    sed "s/{{VERSION}}/${VERSION}/g" |
    sed "s/{{CHANGELOG}}/${ESCAPED_CHANGELOG}/g"
)

# Write the final release body
echo "$RELEASE_BODY" > "$OUTPUT_PATH"