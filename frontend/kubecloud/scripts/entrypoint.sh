#!/bin/sh
set -e

# Escape single quotes in env vars
export VITE_API_BASE_URL=$(printf '%s' "$VITE_API_BASE_URL" | sed "s/'/\\'/g")
export VITE_NETWORK=$(printf '%s' "$VITE_NETWORK" | sed "s/'/\\'/g")
export VITE_STRIPE_PUBLISHABLE_KEY=$(printf '%s' "$VITE_STRIPE_PUBLISHABLE_KEY" | sed "s/'/\\'/g")

# Replace variables in env.js.template with environment values
if [ -f /usr/share/nginx/html/env.js.template ]; then
  envsubst < /usr/share/nginx/html/env.js.template > /usr/share/nginx/html/env.js
fi

exec "$@" 