#!/bin/bash
# Test script for LunaSentri

echo "🧪 Testing LunaSentri Setup"
echo "=========================="

# Check prerequisites
echo "📋 Checking Prerequisites..."

# Check Go
if command -v go &> /dev/null; then
    echo "✅ Go $(go version | awk '{print $3}')"
else
    echo "❌ Go is not installed"
    echo "   Install with: brew install go"
    exit 1
fi

# Check Node.js
if command -v node &> /dev/null; then
    echo "✅ Node.js $(node --version)"
else
    echo "❌ Node.js is not installed"
    exit 1
fi

# Check pnpm
if command -v pnpm &> /dev/null; then
    echo "✅ pnpm $(pnpm --version)"
else
    echo "❌ pnpm is not installed"
    echo "   Install with: npm install -g pnpm"
    exit 1
fi

echo ""
echo "🚀 Starting Tests..."

# Test backend build
echo "🔨 Testing Backend Build..."
cd apps/api-go
if go build -o test-binary .; then
    echo "✅ Backend builds successfully"
    rm -f test-binary
else
    echo "❌ Backend build failed"
    exit 1
fi

# Test frontend build
echo "🔨 Testing Frontend Build..."
cd ../..
if pnpm build:web; then
    echo "✅ Frontend builds successfully"
else
    echo "❌ Frontend build failed"
    exit 1
fi

echo ""
echo "🎉 All tests passed!"
echo ""
echo "🚀 To start development:"
echo "   1. Backend:  cd apps/api-go && go run main.go"
echo "   2. Frontend: pnpm dev:web"
echo ""
echo "🌐 URLs:"
echo "   - Backend:  http://localhost:8080"
echo "   - Frontend: http://localhost:3000"