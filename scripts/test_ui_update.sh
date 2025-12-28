#!/bin/bash

# Usage: ./test_ui_update.sh <SESSION_UUID>
SESSION_ID=$1

if [ -z "$SESSION_ID" ]; then
  echo "Usage: ./test_ui_update.sh <SESSION_UUID>"
  exit 1
fi

echo "🚀 Starting Test Simulation for Session: $SESSION_ID"

echo " -> Test 1: Battery Check (Running)"
docker exec mock-orchestra-mosquitto-1 mosquitto_pub -h localhost \
  -t "tests/$SESSION_ID/result" \
  -m "{\"test_name\": \"Battery Check\", \"status\": \"Running\", \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%S)\"}"

sleep 2

echo " -> Test 1: Battery Check (Passed)"
docker exec mock-orchestra-mosquitto-1 mosquitto_pub -h localhost \
  -t "tests/$SESSION_ID/result" \
  -m "{\"test_name\": \"Battery Check\", \"status\": \"Passed\", \"metrics\": {\"voltage\": 12.5}, \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%S)\"}"

sleep 1

echo " -> Test 2: WiFi Connectivity (Failed)"
docker exec mock-orchestra-mosquitto-1 mosquitto_pub -h localhost \
  -t "tests/$SESSION_ID/result" \
  -m "{\"test_name\": \"WiFi Connectivity\", \"status\": \"Failed\", \"metrics\": {\"signal\": -80, \"error\": \"Timeout\"}, \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%S)\"}"

echo "✅ Simulation Complete. Check UI Table."
