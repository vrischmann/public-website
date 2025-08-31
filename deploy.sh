#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DEPLOY_HOST=${DEPLOY_HOST:-"wevo.rischmann.fr"}
DEPLOY_USER=${DEPLOY_USER:-"your-username"}
DEPLOY_PATH=${DEPLOY_PATH:-"/data/website/"}
BUILD_DIR=${BUILD_DIR:-"build"}

echo -e "${GREEN}Starting deployment...${NC}"

# Check if build directory exists
if [ ! -d "$BUILD_DIR" ]; then
    echo -e "${RED}Error: Build directory '$BUILD_DIR' not found. Run 'just build' first.${NC}"
    exit 1
fi

# Check if SSH key exists
if [ ! -f ~/.ssh/id_rsa ]; then
    echo -e "${YELLOW}Warning: SSH key not found at ~/.ssh/id_rsa${NC}"
    echo -e "${YELLOW}Make sure you have SSH access to $DEPLOY_HOST${NC}"
fi

# Test SSH connection
echo -e "${GREEN}Testing SSH connection...${NC}"
if ! ssh -o ConnectTimeout=10 -o BatchMode=yes "$DEPLOY_USER@$DEPLOY_HOST" "echo 'SSH connection successful'"; then
    echo -e "${RED}Error: Cannot connect to $DEPLOY_HOST${NC}"
    echo -e "${RED}Make sure you have SSH access and the correct DEPLOY_HOST/DEPLOY_USER${NC}"
    exit 1
fi

# Deploy using rsync
echo -e "${GREEN}Deploying to $DEPLOY_HOST:$DEPLOY_PATH...${NC}"
rsync -avz --delete --exclude='.git' "$BUILD_DIR/" "$DEPLOY_USER@$DEPLOY_HOST:$DEPLOY_PATH"

# Verify deployment
echo -e "${GREEN}Verifying deployment...${NC}"
if ssh "$DEPLOY_USER@$DEPLOY_HOST" "ls -la $DEPLOY_PATH | head -10"; then
    echo -e "${GREEN}Deployment successful!${NC}"
    echo -e "${GREEN}Your website should be live at https://$DEPLOY_HOST${NC}"
else
    echo -e "${RED}Warning: Could not verify deployment${NC}"
fi