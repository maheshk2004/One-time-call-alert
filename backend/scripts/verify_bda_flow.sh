#!/bin/bash
set -e

BASE_URL="http://localhost:8085/api"

echo "=========================================================="
echo "Starting Sales BDA Calling Workflow Verification"
echo "=========================================================="

# 1. Login as Sales BDA Vikram
echo "Step 1: Authenticating as Sales BDA Vikram..."
LOGIN_RES=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"vikram@leadfollowup.com","password":"password123"}')
TOKEN=$(echo "$LOGIN_RES" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "FAILED: Could not authenticate as BDA. Response: $LOGIN_RES"
  exit 1
fi
echo "SUCCESS: Authenticated as Vikram. Token acquired."

TIMESTAMP=$(date +%s)

# -------------------------------------------------------------------
# Test 1: 1-Call Unattended -> Moves to unattended_1
# -------------------------------------------------------------------
echo ""
echo "=== Test 1: Student Misses 1st Call (Unattended 1) ==="
LEAD_1=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Admissions Prospect $TIMESTAMP\",\"phone\":\"+91 98888${TIMESTAMP: -5}\",\"email\":\"student_${TIMESTAMP}@example.com\",\"source\":\"College Seminar\",\"course\":\"Full-Stack Web Dev\",\"priority\":\"HIGH\"}")
LEAD_1_ID=$(echo "$LEAD_1" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "1. Created Student Lead: $LEAD_1_ID"

# Log Call Attempt #1 as NO_ANSWER
CALL_1=$(curl -s -X POST "$BASE_URL/leads/$LEAD_1_ID/calls" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"callStatus\":\"NO_ANSWER\",\"callOutcome\":\"FOLLOW_UP_REQUIRED\",\"durationSeconds\":15,\"notes\":\"Attempt 1: Phone ringing, student did not answer\"}")
echo "2. Logged Call Attempt #1 (No Answer)"

# Verify BDA Queue categorizes as unattended_1
QUEUE_1=$(curl -s -X GET "$BASE_URL/bda/queue?category=unattended_1" -H "Authorization: Bearer $TOKEN")
FOUND_1=$(echo "$QUEUE_1" | grep -o "$LEAD_1_ID" || true)
if [ -n "$FOUND_1" ]; then
  echo "SUCCESS: Lead $LEAD_1_ID verified in 'unattended_1' (Missed Call #1) queue!"
else
  echo "FAILED: Lead not found in unattended_1 queue"
  exit 1
fi

# -------------------------------------------------------------------
# Test 2: 2-Calls Unattended -> Moves to unattended_2
# -------------------------------------------------------------------
echo ""
echo "=== Test 2: Student Misses 2nd Call (Unattended 2) ==="
CALL_2=$(curl -s -X POST "$BASE_URL/leads/$LEAD_1_ID/calls" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"callStatus\":\"BUSY\",\"callOutcome\":\"FOLLOW_UP_REQUIRED\",\"durationSeconds\":10,\"notes\":\"Attempt 2: Line busy/switched off\"}")
echo "1. Logged Call Attempt #2 (Busy)"

# Verify BDA Queue categorizes as unattended_2
QUEUE_2=$(curl -s -X GET "$BASE_URL/bda/queue?category=unattended_2" -H "Authorization: Bearer $TOKEN")
FOUND_2=$(echo "$QUEUE_2" | grep -o "$LEAD_1_ID" || true)
if [ -n "$FOUND_2" ]; then
  echo "SUCCESS: Lead $LEAD_1_ID verified in 'unattended_2' (Missed Call #2) queue!"
else
  echo "FAILED: Lead not found in unattended_2 queue"
  exit 1
fi

# -------------------------------------------------------------------
# Test 3: Call Me Later Callback Requested
# -------------------------------------------------------------------
echo ""
echo "=== Test 3: Student Answers and Says 'Call Me Later' ==="
DUE_TIME=$(date -d "+2 hours" --iso-8601=seconds 2>/dev/null || date -v+2H -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date +"%Y-%m-%dT%H:%M:%SZ")
CALL_3=$(curl -s -X POST "$BASE_URL/leads/$LEAD_1_ID/calls" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"callStatus\":\"ANSWERED\",\"callOutcome\":\"CALL_BACK_LATER\",\"durationSeconds\":45,\"nextFollowUpAt\":\"$DUE_TIME\",\"notes\":\"Student asked to call after work at 6pm\"}")
echo "1. Logged Call with CALL_BACK_LATER"

QUEUE_3=$(curl -s -X GET "$BASE_URL/bda/queue?category=callback_later" -H "Authorization: Bearer $TOKEN")
FOUND_3=$(echo "$QUEUE_3" | grep -o "$LEAD_1_ID" || true)
if [ -n "$FOUND_3" ]; then
  echo "SUCCESS: Lead $LEAD_1_ID verified in 'callback_later' (Call Me Later) queue!"
else
  echo "FAILED: Lead not found in callback_later queue"
  exit 1
fi

# -------------------------------------------------------------------
# Test 4: Positive Response -> Interested / Enrolled
# -------------------------------------------------------------------
echo ""
echo "=== Test 4: Student Responds Positively (Interested) ==="
CALL_4=$(curl -s -X POST "$BASE_URL/leads/$LEAD_1_ID/calls" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"callStatus\":\"ANSWERED\",\"callOutcome\":\"INTERESTED\",\"durationSeconds\":200,\"notes\":\"Shared syllabus, candidate wants to enroll!\"}")
echo "1. Logged Call with INTERESTED"

QUEUE_4=$(curl -s -X GET "$BASE_URL/bda/queue?category=interested" -H "Authorization: Bearer $TOKEN")
FOUND_4=$(echo "$QUEUE_4" | grep -o "$LEAD_1_ID" || true)
if [ -n "$FOUND_4" ]; then
  echo "SUCCESS: Lead $LEAD_1_ID verified in 'interested' queue!"
else
  echo "FAILED: Lead not found in interested queue"
  exit 1
fi

# -------------------------------------------------------------------
# Test 5: Not Interested -> Completely Dropped from BDA Queue
# -------------------------------------------------------------------
echo ""
echo "=== Test 5: Student Says Not Interested (Complete Queue Drop) ==="
LEAD_DROP=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Drop Candidate $TIMESTAMP\",\"phone\":\"+91 97777${TIMESTAMP: -5}\",\"email\":\"drop_${TIMESTAMP}@example.com\",\"source\":\"Facebook Ads\",\"course\":\"Data Science\"}")
LEAD_DROP_ID=$(echo "$LEAD_DROP" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "1. Created Candidate to drop: $LEAD_DROP_ID"

CALL_DROP=$(curl -s -X POST "$BASE_URL/leads/$LEAD_DROP_ID/calls" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"callStatus\":\"ANSWERED\",\"callOutcome\":\"NOT_INTERESTED\",\"durationSeconds\":30,\"notes\":\"Candidate chose another college, not interested\"}")
echo "2. Logged Call with NOT_INTERESTED"

# Verify lead is ABSENT from BDA calling queue
QUEUE_ACTIVE=$(curl -s -X GET "$BASE_URL/bda/queue?category=all" -H "Authorization: Bearer $TOKEN")
FOUND_DROP=$(echo "$QUEUE_ACTIVE" | grep -o "$LEAD_DROP_ID" || true)
if [ -z "$FOUND_DROP" ]; then
  echo "SUCCESS: Lead $LEAD_DROP_ID is PERMANENTLY SUPPRESSED from the BDA Calling Queue!"
else
  echo "FAILED: Lead is still appearing in active calling queue"
  exit 1
fi

echo ""
echo "=========================================================="
echo "ALL SALES BDA CALLING WORKFLOWS VERIFIED SUCCESSFULLY (5/5)!"
echo "=========================================================="
