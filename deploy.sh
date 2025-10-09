#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}LunaSentri CapRover Deployment Script${NC}"
echo ""

# Check if caprover CLI is installed
if ! command -v caprover &> /dev/null; then
    echo -e "${RED}CapRover CLI not found. Install with: npm install -g caprover${NC}"
    exit 1
fi

# Parse arguments
SERVICE=$1
if [ -z "$SERVICE" ]; then
    echo -e "${YELLOW}Usage: ./deploy.sh [backend|frontend|all]${NC}"
    exit 1
fi

# Deploy backend
deploy_backend() {
    echo -e "${GREEN}Deploying Backend...${NC}"

    # Create temporary directory
    TEMP_DIR=$(mktemp -d)

    # Copy all backend files to temp directory root
    cp -r apps/api-go/* "$TEMP_DIR/"
    cp deploy/caprover/backend/captain-definition "$TEMP_DIR/"

    # Create tarball from temp directory
    cd "$TEMP_DIR"
    tar -czf deploy.tar.gz *

    # Move tarball to deploy directory
    mv deploy.tar.gz /Users/emiliancon/Desktop/lunasentri/deploy/caprover/backend/

    # Deploy
    cd /Users/emiliancon/Desktop/lunasentri/deploy/caprover/backend
    caprover deploy -t ./deploy.tar.gz

    # Cleanup
    rm -rf "$TEMP_DIR"

    echo -e "${GREEN}Backend deployed successfully!${NC}"
}

# Deploy frontend
deploy_frontend() {
    echo -e "${GREEN}Deploying Frontend...${NC}"

    # Create temporary directory
    TEMP_DIR=$(mktemp -d)

    # Copy only necessary frontend files (exclude node_modules, .next, etc.)
    rsync -av --exclude='node_modules' --exclude='.next' --exclude='.turbo' --exclude='coverage' --exclude='pnpm-workspace.yaml' apps/web-next/ "$TEMP_DIR/"

    # Copy pnpm-lock.yaml from monorepo root
    cp pnpm-lock.yaml "$TEMP_DIR/"
    cp deploy/caprover/frontend/captain-definition "$TEMP_DIR/"

    # Remove workspace file if it was accidentally copied
    rm -f "$TEMP_DIR/pnpm-workspace.yaml"

    # Create tarball from temp directory
    cd "$TEMP_DIR"
    tar -czf deploy.tar.gz *

    # Move tarball to deploy directory
    mv deploy.tar.gz /Users/emiliancon/Desktop/lunasentri/deploy/caprover/frontend/

    # Deploy
    cd /Users/emiliancon/Desktop/lunasentri/deploy/caprover/frontend
    caprover deploy -t ./deploy.tar.gz

    # Cleanup
    rm -rf "$TEMP_DIR"

    echo -e "${GREEN}Frontend deployed successfully!${NC}"
}

# Execute deployment
case $SERVICE in
    backend)
        deploy_backend
        ;;
    frontend)
        deploy_frontend
        ;;
    all)
        deploy_backend
        echo ""
        deploy_frontend
        ;;
    *)
        echo -e "${RED}Invalid service: $SERVICE${NC}"
        echo -e "${YELLOW}Usage: ./deploy.sh [backend|frontend|all]${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}Deployment complete!${NC}"
echo -e "${YELLOW}Don't forget to:${NC}"
echo "  1. Set environment variables in CapRover dashboard"
echo "  2. Enable HTTPS for both apps"
echo "  3. Configure persistent storage for backend (/app/data)"
echo "  4. Set up custom domains"
