#!/bin/sh
set -e

# Update CA certificates if custom certificates are mounted at runtime
if [ -d /usr/local/share/ca-certificates ] && [ "$(ls -A /usr/local/share/ca-certificates 2>/dev/null)" ]; then
    echo "Custom certificates detected, updating CA certificates..."
    update-ca-certificates || echo "WARNING: Failed to update CA certificates"
fi

exec "$@"
