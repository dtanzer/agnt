#!/bin/bash
set -uo pipefail

AGNT="./agnt"
FAKE_TMUX="TMUX=fake PATH=$(pwd)/testtools/bin:$PATH"
WORKSPACE="$(pwd)/test-workspaces/attach-alice.yaml"
SERVER_PID=""

PASS=0
FAIL=0

pass() { echo "PASS: $1"; ((PASS++)); }
fail() { echo "FAIL: $1"; ((FAIL++)); }

check() {
  local desc="$1"
  local expected="$2"
  local actual="$3"
  if echo "$actual" | grep -qF "$expected"; then
    pass "$desc"
  else
    fail "$desc — expected to find: '$expected', got: '$actual'"
  fi
}

check_exit() {
  local desc="$1"
  local expected="$2"
  local actual="$3"
  if [ "$actual" -eq "$expected" ]; then
    pass "$desc"
  else
    fail "$desc — expected exit $expected, got $actual"
  fi
}

TEST_PORT=17717

start_server() {
  local cfg="${1:-$WORKSPACE}"
  TMUX=fake PATH="$(pwd)/testtools/bin:$PATH" $AGNT --workspace-config "$cfg" server start --port $TEST_PORT &>/tmp/agnt-server-test.log &
  SERVER_PID=$!
  sleep 0.5
}

stop_server() {
  if [ -n "$SERVER_PID" ]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
    SERVER_PID=""
  fi
}

cleanup() {
  stop_server
}
trap cleanup EXIT

make -s

echo ""
echo "=== agnt attach acceptance tests ==="
echo ""

# 1. Outside tmux session
output=$(TMUX="" $AGNT --workspace-config $WORKSPACE attach --podman test-abc 2>&1 || true)
exitcode=$(TMUX="" $AGNT --workspace-config $WORKSPACE attach --podman test-abc 2>&1; echo $?)
check "no tmux: error message" "agnt attach must be run inside a tmux session" "$output"
output_exit=$(TMUX="" $AGNT --workspace-config $WORKSPACE attach --podman test-abc 2>&1; echo "EXIT:$?")
check_exit "no tmux: non-zero exit" 1 $(TMUX="" $AGNT --workspace-config $WORKSPACE attach --podman test-abc > /dev/null 2>&1; echo $?)

# 2. No server running
output=$(eval "$FAKE_TMUX $AGNT --workspace-config $WORKSPACE attach --podman test-abc 2>&1" || true)
check "no server: error message" "agnt attach requires a running server" "$output"
check_exit "no server: non-zero exit" 1 $(eval "$FAKE_TMUX $AGNT --workspace-config $WORKSPACE attach --podman test-abc > /dev/null 2>&1"; echo $?)

# 3. Pane with no registered agent (alice-bob workspace has agents on pane 0.2, fake tmux is 0.0)
start_server
output=$(TMUX=fake PATH="$(pwd)/testtools/bin:$PATH" $AGNT --workspace-config $WORKSPACE attach --podman test-abc 2>&1 || true)
# Temporarily use a workspace where no agent is registered to pane 0.0
WORKSPACE_NO_MATCH="$(pwd)/test-workspaces/simple-alice-bob.yaml"
# Start a second server instance for this workspace is complex; instead test with a fresh workspace
# that has no agent on pane 0.0 — use simple-alice-bob which has agents on pane 0.2
stop_server
TMUX=fake PATH="$(pwd)/testtools/bin:$PATH" $AGNT --workspace-config $WORKSPACE_NO_MATCH server start --port $TEST_PORT &>/tmp/agnt-server-test2.log &
SERVER_PID=$!
sleep 0.5
output=$(TMUX=fake PATH="$(pwd)/testtools/bin:$PATH" $AGNT --workspace-config $WORKSPACE_NO_MATCH attach --podman test-abc 2>&1 || true)
check "unknown pane: error message" "no agent registered to the current tmux pane" "$output"
check_exit "unknown pane: non-zero exit" 1 $(TMUX=fake PATH="$(pwd)/testtools/bin:$PATH" $AGNT --workspace-config $WORKSPACE_NO_MATCH attach --podman test-abc > /dev/null 2>&1; echo $?)
stop_server

# 4. server status with no attached agents
start_server
output=$(eval "$FAKE_TMUX $AGNT --workspace-config $WORKSPACE server status 2>&1")
check "status no agents: shows 'Agents:   none'" "Agents:   none" "$output"
stop_server

# 5. Successful attach
start_server
output=$(eval "$FAKE_TMUX $AGNT --workspace-config $WORKSPACE attach --podman alice-x3k2m9f7 2>&1")
check "attach success: output message" "Attached alice (podman: alice-x3k2m9f7)" "$output"
check_exit "attach success: exit 0" 0 $(eval "$FAKE_TMUX $AGNT --workspace-config $WORKSPACE attach --podman alice-x3k2m9f7 > /dev/null 2>&1"; echo $?)

# 6. Server status shows attached agent
output=$(eval "$FAKE_TMUX $AGNT --workspace-config $WORKSPACE server status 2>&1")
check "status after attach: shows agent" "alice (podman: alice-x3k2m9f7)" "$output"

# 7. Attach same name twice — replaces entry
eval "$FAKE_TMUX $AGNT --workspace-config $WORKSPACE attach --podman alice-newcontainer > /dev/null 2>&1" || true
output=$(eval "$FAKE_TMUX $AGNT --workspace-config $WORKSPACE server status 2>&1")
check "re-attach: new container name shown" "alice (podman: alice-newcontainer)" "$output"

# 8. Restart server clears state
stop_server
start_server
output=$(eval "$FAKE_TMUX $AGNT --workspace-config $WORKSPACE server status 2>&1")
check "after restart: agents cleared" "Agents:   none" "$output"
stop_server

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
echo ""
[ "$FAIL" -eq 0 ]
