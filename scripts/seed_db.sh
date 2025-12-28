#!/bin/bash

# Configuration
# 1. Name of your MongoDB service in docker-compose (usually 'mongo' or 'db')
MONGO_CONTAINER="mock-orchestra-mongo-1" 
# 2. Name of your Database
DB_NAME="test_hub_db"
# 3. Name of the Collection
COLLECTION="build_numbers"

echo "🌱 Seeding MongoDB ($MONGO_CONTAINER) with test build numbers..."

# Check if container is running
if ! docker ps | grep -q "$MONGO_CONTAINER"; then
    echo "❌ Error: Container '$MONGO_CONTAINER' is not running."
    echo "   Did you mean 'mongo' instead of 'mongo-1'? Check 'docker ps'."
    exit 1
fi

# The Javascript command to execute inside MongoDB
# We use updateOne with upsert:true to avoid duplicates if you run this twice.
JS_CMD="
db = db.getSiblingDB('$DB_NAME');
var builds = ['1.0.0-alpha', '1.0.0-beta', '2.0.0-rc1'];

builds.forEach(function(b) {
  db.$COLLECTION.updateOne(
    { build_number: b }, 
    { \$set: { build_number: b, created_at: new Date() } }, 
    { upsert: true }
  );
  print('✅ Inserted/Updated: ' + b);
});
"

# Execute via docker exec
# We assume 'mongosh' exists (Mongo 5+). If using old Mongo, change to 'mongo'.
docker exec -i "$MONGO_CONTAINER" mongosh --eval "$JS_CMD"

echo "✨ Database seeding complete."

