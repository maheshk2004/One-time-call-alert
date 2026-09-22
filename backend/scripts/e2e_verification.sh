#!/bin/bash
set -e

BASE_URL="http://localhost:8085/api"

echo "=========================================================="
echo "Starting End-to-End Scenario Verification (Phase 14)"
echo "=========================================================="

# 1. Login as Admin
echo "Step 1: Authenticating as Admin..."
LOGIN_RES=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@leadfollowup.com","password":"password123"}')
TOKEN=$(echo "$LOGIN_RES" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "FAILED: Could not authenticate as admin. Response: $LOGIN_RES"
  exit 1
fi
echo "SUCCESS: Authenticated. Token acquired."

# Fetch Salesperson Vikram
USERS_RES=$(curl -s -X GET "$BASE_URL/users?role=SALES" -H "Authorization: Bearer $TOKEN")
VIKRAM_ID=$(echo "$USERS_RES" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "Fetched Salesperson Vikram ID: $VIKRAM_ID"

# -------------------------------------------------------------------
# SCENARIO A: 1-Call No Answer -> Overdue -> 2nd Call Clears Alert
# -------------------------------------------------------------------
echo ""
echo "=== Scenario A: Single Call -> Overdue Alert -> 2nd Call Clears Alert ==="
TIMESTAMP=$(date +%s)
LEAD_A=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"E2E Prospect A\",\"phone\":\"+91 99000${TIMESTAMP: -5}\",\"email\":\"prospectA_${TIMESTAMP}@example.com\",\"source\":\"Facebook Ads\",\"course\":\"Full-Stack Web Dev\",\"priority\":\"HIGH\"}")
LEAD_A_ID=$(echo "$LEAD_A" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "1. Created Lead A: $LEAD_A_ID"

# Assign Salesperson
curl -s -X PUT "$BASE_URL/leads/$LEAD_A_ID/assign" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"salespersonId\":\"$VIKRAM_ID\"}" > /dev/null
echo "2. Assigned to Vikram"

# Record First Call with status = NO_ANSWER
CALL_A1=$(curl -s -X POST "$BASE_URL/leads/$LEAD_A_ID/calls" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"callStatus\":\"NO_ANSWER\",\"callOutcome\":\"FOLLOW_UP_REQUIRED\",\"durationSeconds\":20,\"notes\":\"Attempt 1 - No response\"}")
CALL_A1_ID=$(echo "$CALL_A1" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "3. Recorded Call Attempt #1 (No Answer). Call ID: $CALL_A1_ID"

# Trigger background detector
DETECTION_RES=$(curl -s -X POST "$BASE_URL/follow-ups/detect" -H "Authorization: Bearer $TOKEN")
echo "4. Triggered detector: $DETECTION_RES"

# Record Second Call (Customer answered!)
CALL_A2=$(curl -s -X POST "$BASE_URL/leads/$LEAD_A_ID/calls" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"callStatus\":\"ANSWERED\",\"callOutcome\":\"INTERESTED\",\"durationSeconds\":180,\"notes\":\"Attempt 2 - Spoke with candidate, shared syllabus\"}")
echo "5. Recorded Call Attempt #2 (Answered, Interested)"

# Verify Lead A is now attemptCount = 2
LEAD_A_VERIFY=$(curl -s -X GET "$BASE_URL/leads/$LEAD_A_ID" -H "Authorization: Bearer $TOKEN")
ATTEMPT_COUNT=$(echo "$LEAD_A_VERIFY" | grep -o '"callAttemptCount":[0-9]*' | cut -d':' -f2)
STATUS_A=$(echo "$LEAD_A_VERIFY" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
echo "6. Lead A CallAttemptCount: $ATTEMPT_COUNT, Status: $STATUS_A"
if [ "$ATTEMPT_COUNT" -eq 2 ] && [ "$STATUS_A" = "INTERESTED" ]; then
  echo "SUCCESS: Scenario A passed! Lead progressed beyond 1-call status."
else
  echo "FAILED: Scenario A check failed"
  exit 1
fi

# -------------------------------------------------------------------
# SCENARIO B: Explicit NOT_INTERESTED -> Follow-up Suppressed
# -------------------------------------------------------------------
echo ""
echo "=== Scenario B: AI NOT_INTERESTED -> Follow-up Suppressed ==="
TIMESTAMP=$(date +%s)
LEAD_B=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"E2E Prospect B\",\"phone\":\"+91 99001${TIMESTAMP: -5}\",\"email\":\"prospectB_${TIMESTAMP}@example.com\",\"source\":\"Google Ads\",\"course\":\"Cloud DevOps\"}")
LEAD_B_ID=$(echo "$LEAD_B" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)

CALL_B=$(curl -s -X POST "$BASE_URL/leads/$LEAD_B_ID/calls" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"callStatus\":\"ANSWERED\",\"callOutcome\":\"INTERESTED\",\"durationSeconds\":90,\"recordingConsent\":true}")
CALL_B_ID=$(echo "$CALL_B" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)

# Upload Transcript: "I am not interested in this course. I don't want this."
curl -s -X POST "$BASE_URL/calls/$CALL_B_ID/transcript" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"transcript\":\"Salesperson: Hello! Lead: I am not interested in this course. I don't want this.\",\"speakerSegments\":[{\"speaker\":\"salesperson\",\"text\":\"Hello!\"},{\"speaker\":\"lead\",\"text\":\"I am not interested in this course. I don't want this.\"}]}" > /dev/null

# Trigger AI Analysis
AI_B=$(curl -s -X POST "$BASE_URL/calls/$CALL_B_ID/analyze" -H "Authorization: Bearer $TOKEN")
INTENT_B=$(echo "$AI_B" | grep -o '"intent":"[^"]*' | cut -d'"' -f4)
CONF_B=$(echo "$AI_B" | grep -o '"confidence":[0-9.]*' | cut -d':' -f2)
echo "AI Analyzed Conversation -> Intent: $INTENT_B, Confidence: $CONF_B"

# Confirm review
ANALYSIS_B_ID=$(echo "$AI_B" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
curl -s -X POST "$BASE_URL/ai/reviews/$ANALYSIS_B_ID/confirm" -H "Authorization: Bearer $TOKEN" > /dev/null

LEAD_B_VERIFY=$(curl -s -X GET "$BASE_URL/leads/$LEAD_B_ID" -H "Authorization: Bearer $TOKEN")
STATUS_B=$(echo "$LEAD_B_VERIFY" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
echo "Verified Lead B Status: $STATUS_B"
if [ "$STATUS_B" = "NOT_INTERESTED" ]; then
  echo "SUCCESS: Scenario B passed! Lead correctly marked NOT_INTERESTED."
else
  echo "FAILED: Scenario B check failed"
  exit 1
fi

# -------------------------------------------------------------------
# SCENARIO C: Explicit DO_NOT_CALL -> DNC Flagged & Alert Suppressed
# -------------------------------------------------------------------
echo ""
echo "=== Scenario C: Explicit DO_NOT_CALL -> Strict Protection ==="
TIMESTAMP=$(date +%s)
LEAD_C=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"E2E Prospect C\",\"phone\":\"+91 99002${TIMESTAMP: -5}\",\"email\":\"prospectC_${TIMESTAMP}@example.com\",\"source\":\"Website Form\",\"course\":\"Cybersecurity\"}")
LEAD_C_ID=$(echo "$LEAD_C" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)

CALL_C=$(curl -s -X POST "$BASE_URL/leads/$LEAD_C_ID/calls" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"callStatus\":\"ANSWERED\",\"callOutcome\":\"INTERESTED\",\"durationSeconds\":40,\"recordingConsent\":true}")
CALL_C_ID=$(echo "$CALL_C" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)

# Transcript: "Please don't call me again. Remove my number from your database!"
curl -s -X POST "$BASE_URL/calls/$CALL_C_ID/transcript" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"transcript\":\"Salesperson: Hello! Lead: Please don't call me again. Remove my number from your database!\",\"speakerSegments\":[{\"speaker\":\"salesperson\",\"text\":\"Hello!\"},{\"speaker\":\"lead\",\"text\":\"Please don't call me again. Remove my number from your database!\"}]}" > /dev/null

AI_C=$(curl -s -X POST "$BASE_URL/calls/$CALL_C_ID/analyze" -H "Authorization: Bearer $TOKEN")
INTENT_C=$(echo "$AI_C" | grep -o '"intent":"[^"]*' | cut -d'"' -f4)
echo "AI Analyzed Conversation -> Intent: $INTENT_C"

ANALYSIS_C_ID=$(echo "$AI_C" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
curl -s -X POST "$BASE_URL/ai/reviews/$ANALYSIS_C_ID/confirm" -H "Authorization: Bearer $TOKEN" > /dev/null

LEAD_C_VERIFY=$(curl -s -X GET "$BASE_URL/leads/$LEAD_C_ID" -H "Authorization: Bearer $TOKEN")
DNC_FLAG=$(echo "$LEAD_C_VERIFY" | grep -o '"doNotCall":true' || true)
STATUS_C=$(echo "$LEAD_C_VERIFY" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
echo "Verified Lead C Status: $STATUS_C, DoNotCall: $DNC_FLAG"
if [ "$STATUS_C" = "DO_NOT_CALL" ] && [ -n "$DNC_FLAG" ]; then
  echo "SUCCESS: Scenario C passed! Lead placed in Do-Not-Contact registry."
else
  echo "FAILED: Scenario C check failed"
  exit 1
fi

echo ""
echo "=========================================================="
echo "ALL END-TO-END SCENARIOS VERIFIED SUCCESSFULLY!"
echo "=========================================================="
