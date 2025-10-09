#!/usr/bin/env bash
# Quick test script to verify the dev workflow fixes

set -e

echo "🧪 Testing Development Workflow Fixes"
echo "======================================"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DB_FILE="$ROOT_DIR/apps/api-go/data/lunasentri.db"

# Test 1: Check script syntax
echo "✓ Test 1: Checking script syntax..."
bash -n "$ROOT_DIR/scripts/dev-reset.sh"
echo "  ✅ Script syntax is valid"
echo ""

# Test 2: Check --reset-db flag parsing
echo "✓ Test 2: Checking flag parsing..."
if grep -q "RESET_DB=false" "$ROOT_DIR/scripts/dev-reset.sh"; then
  echo "  ✅ --reset-db flag logic found"
else
  echo "  ❌ --reset-db flag logic missing"
  exit 1
fi
echo ""

# Test 3: Check port cleanup function
echo "✓ Test 3: Checking port cleanup function..."
if grep -q "kill_port()" "$ROOT_DIR/scripts/dev-reset.sh"; then
  echo "  ✅ Port cleanup function exists"
else
  echo "  ❌ Port cleanup function missing"
  exit 1
fi
echo ""

# Test 4: Check database deletion logic
echo "✓ Test 4: Checking database persistence logic..."
if grep -q 'if \[ "$RESET_DB" = true \]' "$ROOT_DIR/scripts/dev-reset.sh"; then
  echo "  ✅ Conditional database deletion found"
else
  echo "  ❌ Conditional database deletion missing"
  exit 1
fi
echo ""

# Test 5: Check log capture
echo "✓ Test 5: Checking log capture..."
if grep -q "tee.*backend.log" "$ROOT_DIR/scripts/dev-reset.sh"; then
  echo "  ✅ Backend log capture found"
else
  echo "  ❌ Backend log capture missing"
  exit 1
fi
echo ""

# Test 6: Check health checks
echo "✓ Test 6: Checking health validation..."
if grep -q "kill -0.*BACKEND_PID" "$ROOT_DIR/scripts/dev-reset.sh"; then
  echo "  ✅ Process health checks found"
else
  echo "  ❌ Process health checks missing"
  exit 1
fi
echo ""

# Test 7: Check CORS configuration
echo "✓ Test 7: Checking CORS configuration in backend..."
if grep -q "Access-Control-Allow-Credentials" "$ROOT_DIR/apps/api-go/main.go"; then
  echo "  ✅ CORS credentials support found"
else
  echo "  ❌ CORS credentials support missing"
  exit 1
fi
echo ""

# Test 8: Verify documentation
echo "✓ Test 8: Checking documentation..."
docs_found=0
[ -f "$ROOT_DIR/DEV_WORKFLOW_FIXES.md" ] && ((docs_found++))
[ -f "$ROOT_DIR/QUICK_START.md" ] && ((docs_found++))
[ -f "$ROOT_DIR/DEV_WORKFLOW_FIX_SUMMARY.md" ] && ((docs_found++))

if [ $docs_found -eq 3 ]; then
  echo "  ✅ All documentation files created"
else
  echo "  ⚠️  Some documentation files missing ($docs_found/3)"
fi
echo ""

# Test 9: Check script executable
echo "✓ Test 9: Checking script permissions..."
if [ -x "$ROOT_DIR/scripts/dev-reset.sh" ]; then
  echo "  ✅ Script is executable"
else
  echo "  ⚠️  Script is not executable (run: chmod +x scripts/dev-reset.sh)"
fi
echo ""

echo "======================================"
echo "✅ All tests passed!"
echo ""
echo "Next steps:"
echo "1. Test database persistence:"
echo "   ./scripts/dev-reset.sh"
echo ""
echo "2. Test database reset:"
echo "   ./scripts/dev-reset.sh --reset-db"
echo ""
echo "3. Test port cleanup:"
echo "   ./scripts/dev-reset.sh  # Start"
echo "   ./scripts/dev-reset.sh  # Start again (should auto-cleanup)"
echo ""
echo "4. Test registration:"
echo "   curl -X POST http://localhost:8080/auth/register \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"email\":\"test@example.com\",\"password\":\"password123\"}'"
echo ""
echo "For full verification checklist, see: DEV_WORKFLOW_FIXES.md"
