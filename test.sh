#!/bin/bash

# Test script for SSH Manager
echo "=== SSH Manager Test Suite ==="
echo ""

# Test 1: Help
echo "Test 1: Help command"
./sshmgr --help > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ Help command works"
else
    echo "✗ Help command failed"
fi
echo ""

# Test 2: Init (create test config backup)
echo "Test 2: Initialize (test with short password)"
echo "short" | ./sshmgr init > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "✓ Password validation works (rejected short password)"
else
    echo "✗ Password validation failed"
fi
echo ""

# Test 3: List (empty)
echo "Test 3: List hosts (empty)"
output=$(./sshmgr list 2>&1)
if echo "$output" | grep -q "No hosts found"; then
    echo "✓ List command works (empty state)"
else
    echo "✗ List command failed"
fi
echo ""

# Test 4: Add (without master password)
echo "Test 4: Add host (should fail - no master password)"
echo -e "test\nhost\nuser\npass\n22\nn" | ./sshmgr add > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "✓ Add command requires authentication"
else
    echo "✗ Add command failed to require authentication"
fi
echo ""

echo "=== All basic tests completed ==="
echo ""
echo "Note: Full integration tests require:"
echo "  - Running 'sshmgr init' with real password"
echo "  - Adding hosts with 'sshmgr add'"
echo "  - Connecting with 'sshmgr connect <alias>'"
echo ""
echo "Binary location: $(pwd)/sshmgr"
echo "Binary size: $(du -h sshmgr | cut -f1)"
