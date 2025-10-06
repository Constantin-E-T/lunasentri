#!/bin/bash

echo "=== Register first user (should become admin) ==="
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"SecurePass123"}'
echo ""

echo ""
echo "=== Register second user (should be non-admin) ==="
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"AnotherPass123"}'
echo ""

echo ""
echo "=== Login with first user (should show is_admin: true) ==="
curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"SecurePass123"}' \
  -c /tmp/cookies.txt
echo ""

echo ""
echo "=== Login with second user (should show is_admin: false) ==="
curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"AnotherPass123"}' \
  -c /tmp/cookies2.txt
echo ""

echo ""
echo "=== Test validation: password too short ==="
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"short"}'
echo ""

echo ""
echo "=== Test validation: invalid email ==="
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"notanemail","password":"LongPassword123"}'
echo ""
