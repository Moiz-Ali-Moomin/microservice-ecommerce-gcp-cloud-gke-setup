#!/bin/bash
set -e

METABASE_URL="http://metabase.tooling-metabase.svc.cluster.local:3000"
METABASE_API="$METABASE_URL/api"
DB_HOST="postgres-postgresql-primary.data-postgres.svc.cluster.local"
DB_PORT="5432"
DB_NAME="ecommerce_db"

echo "Waiting for Metabase to be ready..."
until curl -s "$METABASE_API/health" | grep -q '{"status":"ok"}'; do
  echo "Metabase is not ready yet. Retrying in 10 seconds..."
  sleep 10
done

echo "Metabase is ready! Authenticating..."

# Admin credentials should ideally be injected via secret, but since Metabase 
# requires an initial setup for the first admin, we'll try to set up or login.
# Assuming standard setup properties or an API token is provided.

# 1. Try to login (if already set up)
LOGIN_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
  -d '{"username": "'"$METABASE_ADMIN_EMAIL"'", "password": "'"$METABASE_ADMIN_PASSWORD"'"}' \
  "$METABASE_API/session")

SESSION_ID=$(echo "$LOGIN_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -z "$SESSION_ID" ]; then
    echo "Login failed. Attempting initial setup..."
    
    SETUP_TOKEN=$(curl -s "$METABASE_API/session/properties" | grep -o '"setup-token":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$SETUP_TOKEN" ]; then
        echo "Error: Could not obtain setup token and login failed. Metabase might already be set up with different credentials."
        exit 1
    fi
    
    SETUP_PAYLOAD=$(cat <<EOF
    {
      "token": "$SETUP_TOKEN",
      "user": {
        "first_name": "Admin",
        "last_name": "User",
        "email": "$METABASE_ADMIN_EMAIL",
        "password": "$METABASE_ADMIN_PASSWORD",
        "site_name": "OpsYield"
      },
      "prefs": {
        "allow_tracking": false,
        "site_name": "OpsYield"
      },
      "database": null
    }
EOF
)
    
    SETUP_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "$SETUP_PAYLOAD" "$METABASE_API/setup")
    SESSION_ID=$(echo "$SETUP_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$SESSION_ID" ]; then
        echo "Error: Initial setup failed. Response: $SETUP_RESPONSE"
        exit 1
    fi
    echo "Initial setup successful."
else
    echo "Login successful."
fi

echo "Adding PostgreSQL database as a data source..."

DB_PAYLOAD=$(cat <<EOF
{
  "name": "Ecommerce DB (ELT)",
  "engine": "postgres",
  "details": {
    "host": "$DB_HOST",
    "port": $DB_PORT,
    "dbname": "$DB_NAME",
    "user": "$DB_USER",
    "password": "$DB_PASSWORD",
    "ssl": false
  },
  "is_full_sync": true,
  "is_on_demand": false,
  "schedules": {
    "cache_field_values": {
      "schedule_day": null,
      "schedule_frame": null,
      "schedule_hour": 0,
      "schedule_type": "daily"
    },
    "metadata_sync": {
      "schedule_day": null,
      "schedule_frame": null,
      "schedule_hour": null,
      "schedule_type": "hourly"
    }
  }
}
EOF
)

# Check if DB already exists
EXISTING_DBS=$(curl -s -H "X-Metabase-Session: $SESSION_ID" "$METABASE_API/database")
if echo "$EXISTING_DBS" | grep -q '"name":"Ecommerce DB (ELT)"'; then
    echo "Database 'Ecommerce DB (ELT)' already exists. Skipping."
else
    CREATE_DB_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "X-Metabase-Session: $SESSION_ID" -d "$DB_PAYLOAD" "$METABASE_API/database")
    
    if echo "$CREATE_DB_RESPONSE" | grep -q '"id"'; then
        echo "Successfully added PostgreSQL data source!"
    else
        echo "Failed to add data source: $CREATE_DB_RESPONSE"
        exit 1
    fi
fi

echo "Metabase provisioning complete."
