#!/bin/bash

# Fix WhatsApp Center Service
# Script untuk memperbaiki systemd service

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔧 WhatsApp Center Service Fixer${NC}"
echo "================================="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}❌ This script must be run as root (use sudo)${NC}"
   exit 1
fi

echo -e "${YELLOW}🔍 Step 1: Stop and disable current service${NC}"
systemctl stop whatsapp-center 2>/dev/null || true
systemctl disable whatsapp-center 2>/dev/null || true

echo -e "${YELLOW}🔍 Step 2: Check current directory structure${NC}"
echo "Current directory: $(pwd)"
echo "Files in current directory:"
ls -la

echo ""
echo -e "${YELLOW}🔍 Step 3: Check if whatsapp binary exists${NC}"
CURRENT_DIR=$(pwd)
BINARY_PATH="$CURRENT_DIR/linux-amd64"
if [ -f "$BINARY_PATH" ]; then
    echo -e "${GREEN}✅ Binary found: $BINARY_PATH${NC}"
    ls -la "$BINARY_PATH"
    file "$BINARY_PATH"
else
    echo -e "${RED}❌ Binary not found: $BINARY_PATH${NC}"
    echo "Available files in current directory:"
    ls -la
    echo "Please check the correct path to your whatsapp binary"
    exit 1
fi

echo ""
echo -e "${YELLOW}🔍 Step 4: Test binary manually${NC}"
echo "Testing binary execution..."
cd "$CURRENT_DIR"
timeout 5s ./linux-amd64 --help || echo "Binary test completed (timeout expected)"

echo ""
echo -e "${YELLOW}🔍 Step 5: Check .env file${NC}"
if [ -f "$CURRENT_DIR/.env" ]; then
    echo -e "${GREEN}✅ .env file found${NC}"
    echo "Environment variables will be loaded from .env file"
else
    echo -e "${YELLOW}⚠️  .env file not found, checking for .env.example${NC}"
    if [ -f "$CURRENT_DIR/.env.example" ]; then
        echo -e "${GREEN}✅ .env.example found, copying to .env${NC}"
        cp "$CURRENT_DIR/.env.example" "$CURRENT_DIR/.env"
        echo -e "${YELLOW}⚠️  Please edit .env file with your actual configuration${NC}"
    else
        echo -e "${RED}❌ No .env or .env.example file found${NC}"
        echo "Creating basic .env file..."
        cat > "$CURRENT_DIR/.env" << 'ENVEOF'
PORT=3000
HOST=0.0.0.0
LOG_LEVEL=info
OTOMAX_ENABLED=true
OTOMAX_API_URL=https://api.otomax.id
OTOMAX_APP_ID=your_app_id
OTOMAX_APP_KEY=your_app_key
OTOMAX_DEV_KEY=your_dev_key
OTOMAX_DEFAULT_RESELLER=default_reseller
OTOMAX_AUTO_REPLY_ENABLED=true
ENVEOF
        echo -e "${YELLOW}⚠️  Basic .env file created. Please edit with your actual configuration${NC}"
    fi
fi

echo ""
echo -e "${YELLOW}🔍 Step 6: Create simple service file${NC}"
cat > /etc/systemd/system/whatsapp-center.service << EOF
[Unit]
Description=WhatsApp Center - WhatsApp Web Multidevice API with OtomaX Integration
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=$CURRENT_DIR
ExecStart=$CURRENT_DIR/linux-amd64 rest
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Resource limits
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

echo -e "${GREEN}✅ Simple service file created${NC}"

echo ""
echo -e "${YELLOW}🔍 Step 7: Reload systemd and enable service${NC}"
systemctl daemon-reload
systemctl enable whatsapp-center

echo ""
echo -e "${YELLOW}🔍 Step 8: Start service${NC}"
systemctl start whatsapp-center

echo ""
echo -e "${YELLOW}🔍 Step 9: Check service status${NC}"
sleep 3
systemctl status whatsapp-center --no-pager -l

echo ""
echo -e "${YELLOW}🔍 Step 10: Show recent logs${NC}"
journalctl -u whatsapp-center --no-pager -l -n 10

echo ""
echo -e "${GREEN}🎉 Service fix completed!${NC}"
echo ""
echo -e "${BLUE}📋 Useful commands:${NC}"
echo "• Check status: systemctl status whatsapp-center"
echo "• View logs: journalctl -u whatsapp-center -f"
echo "• Restart: systemctl restart whatsapp-center"
echo "• Stop: systemctl stop whatsapp-center"
