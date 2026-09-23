#!/bin/bash
set -e

BASE_URL="http://localhost:8085/api"

echo "=========================================================="
echo "Starting AI Dialogue Auto-Decision Engine Verification"
echo "Zero Manual BDA Status Entry Test"
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
# Test 1: AI Broken / Dropped Call Detection
# -------------------------------------------------------------------
echo ""
echo "=== Test 1: Broken Call Transcript (Voice breaking / call cut) ==="
LEAD_1=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Broken Call Student $TIMESTAMP\",\"phone\":\"+91 99111${TIMESTAMP: -5}\",\"email\":\"broken_${TIMESTAMP}@example.com\",\"source\":\"Website Form\",\"course\":\"Full-Stack Python\",\"priority\":\"HIGH\"}")
LEAD_1_ID=$(echo "$LEAD_1" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "1. Created Student Lead: $LEAD_1_ID"

# BDA only submits the spoken transcript text (NO status passed by BDA!)
AUTO_1=$(curl -s -X POST "$BASE_URL/leads/$LEAD_1_ID/auto-call" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"transcriptText\":\"Sales BDA: Hello, am I speaking with student?\\nStudent: Hello? Hello? Your voice is breaking up badly, call cut...\\nSales BDA: Hello, can you hear me?\",\"durationSeconds\":16}")

INTENT_1=$(echo "$AUTO_1" | grep -o '"intent":"[^"]*' | head -n 1 | cut -d'"' -f4)
EVIDENCE_1=$(echo "$AUTO_1" | grep -o '"text":"[^"]*' | head -n 1 | cut -d'"' -f4 || true)
echo "2. AI Auto-Processed Call: Intent = $INTENT_1, Evidence = '$EVIDENCE_1'"

# Check Lead status in database
LEAD_1_CHECK=$(curl -s -X GET "$BASE_URL/leads/$LEAD_1_ID" -H "Authorization: Bearer $TOKEN")
STATUS_1=$(echo "$LEAD_1_CHECK" | grep -o '"status":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "3. Lead Status Automatically Set to: $STATUS_1"

if [ "$STATUS_1" = "CALLED_ONCE" ] || [ "$STATUS_1" = "FOLLOW_UP_REQUIRED" ]; then
  echo "SUCCESS: AI automatically flagged broken call for follow-up without manual BDA status entry!"
else
  echo "FAILED: Unexpected lead status: $STATUS_1"
  exit 1
fi

# -------------------------------------------------------------------
# Test 2: AI Call Back Later Detection
# -------------------------------------------------------------------
echo ""
echo "=== Test 2: 'Call Me Later' Transcript (Student in Meeting) ==="
LEAD_2=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Busy Meeting Student $TIMESTAMP\",\"phone\":\"+91 99222${TIMESTAMP: -5}\",\"email\":\"meeting_${TIMESTAMP}@example.com\",\"source\":\"College Seminar\",\"course\":\"Data Science\",\"priority\":\"URGENT\"}")
LEAD_2_ID=$(echo "$LEAD_2" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "1. Created Student Lead: $LEAD_2_ID"

AUTO_2=$(curl -s -X POST "$BASE_URL/leads/$LEAD_2_ID/auto-call" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"transcriptText\":\"Sales BDA: Hello, calling regarding your data science admission.\\nStudent: Hi, I am currently in an urgent office meeting. Please call me later this evening around 6 PM.\\nSales BDA: Understood, I will call you this evening at 6.\",\"durationSeconds\":32}")

INTENT_2=$(echo "$AUTO_2" | grep -o '"intent":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "2. AI Auto-Processed Call: Intent = $INTENT_2"

LEAD_2_CHECK=$(curl -s -X GET "$BASE_URL/leads/$LEAD_2_ID" -H "Authorization: Bearer $TOKEN")
STATUS_2=$(echo "$LEAD_2_CHECK" | grep -o '"status":"[^"]*' | head -n 1 | cut -d'"' -f4)
FOLLOWUP_2=$(echo "$LEAD_2_CHECK" | grep -o '"nextFollowUpAt":"[^"]*' || true)
echo "3. Lead Status Automatically Set to: $STATUS_2 ($FOLLOWUP_2)"

if [ "$STATUS_2" = "FOLLOW_UP_SCHEDULED" ] && [ -n "$FOLLOWUP_2" ]; then
  echo "SUCCESS: AI automatically scheduled callback and updated status to FOLLOW_UP_SCHEDULED!"
else
  echo "FAILED: Callback was not scheduled automatically"
  exit 1
fi

# -------------------------------------------------------------------
# Test 3: AI Not Interested / Do Not Call Detection
# -------------------------------------------------------------------
echo ""
echo "=== Test 3: Not Interested Transcript (Drop Lead & DNC) ==="
LEAD_3=$(curl -s -X POST "$BASE_URL/leads" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Not Interested Student $TIMESTAMP\",\"phone\":\"+91 99333${TIMESTAMP: -5}\",\"email\":\"notinterested_${TIMESTAMP}@example.com\",\"source\":\"Google Ads\",\"course\":\"Cloud Computing\",\"priority\":\"LOW\"}")
LEAD_3_ID=$(echo "$LEAD_3" | grep -o '"id":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "1. Created Student Lead: $LEAD_3_ID"

AUTO_3=$(curl -s -X POST "$BASE_URL/leads/$LEAD_3_ID/auto-call" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"transcriptText\":\"Sales BDA: Hello, following up on your cloud computing application.\\nStudent: I already took admission in another institute. Not interested at all, please do not call me again.\\nSales BDA: Okay, removing your number from our system.\",\"durationSeconds\":22}")

INTENT_3=$(echo "$AUTO_3" | grep -o '"intent":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "2. AI Auto-Processed Call: Intent = $INTENT_3"

LEAD_3_CHECK=$(curl -s -X GET "$BASE_URL/leads/$LEAD_3_ID" -H "Authorization: Bearer $TOKEN")
STATUS_3=$(echo "$LEAD_3_CHECK" | grep -o '"status":"[^"]*' | head -n 1 | cut -d'"' -f4)
echo "3. Lead Status Automatically Set to: $STATUS_3"

# Verify that this lead is ABSENT from BDA active calling queue
QUEUE_ALL=$(curl -s -X GET "$BASE_URL/bda/queue?category=all" -H "Authorization: Bearer $TOKEN")
FOUND_3=$(echo "$QUEUE_ALL" | grep -o "$LEAD_3_ID" || true)

if [ -z "$FOUND_3" ]; then
  echo "SUCCESS: AI automatically dropped Not-Interested student from BDA Calling Queue!"
else
  echo "FAILED: Student is still present in BDA calling queue"
  exit 1
fi

echo ""
echo "=========================================================="
echo "ALL AI TRANSCRIPT AUTO-DECISION TESTS PASSED (3/3)!"
echo "Zero manual BDA status entry verified."
echo "=========================================================="
