#!/bin/bash

# WhatsApp Center Linux Service Installer
# This script will install WhatsApp Center as a systemd service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="whatsapp-center"
SERVICE_USER="whatsapp"
SERVICE_GROUP="whatsapp"
INSTALL_DIR="/opt/whatsapp-center"
SERVICE_FILE="whatsapp-center.service"

echo -e "${BLUE}🚀 WhatsApp Center Service Installer${NC}"
echo "=================================="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}❌ This script must be run as root (use sudo)${NC}"
   exit 1
fi

# Check if service file exists
if [[ ! -f "$SERVICE_FILE" ]]; then
    echo -e "${RED}❌ Service file $SERVICE_FILE not found in current directory${NC}"
    exit 1
fi

echo -e "${YELLOW}📋 Installation Steps:${NC}"
echo "1. Create user and group"
echo "2. Create installation directory"
echo "3. Copy service file to systemd"
echo "4. Enable and start service"
echo ""

# Step 1: Create user and group
echo -e "${BLUE}👤 Creating user and group...${NC}"
if ! id "$SERVICE_USER" &>/dev/null; then
    useradd --system --no-create-home --shell /bin/false "$SERVICE_USER"
    echo -e "${GREEN}✅ User $SERVICE_USER created${NC}"
else
    echo -e "${YELLOW}⚠️  User $SERVICE_USER already exists${NC}"
fi

# Step 2: Create installation directory
echo -e "${BLUE}📁 Creating installation directory...${NC}"
mkdir -p "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR/storages"
mkdir -p "$INSTALL_DIR/statics"

# Set ownership
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
echo -e "${GREEN}✅ Installation directory created: $INSTALL_DIR${NC}"

# Step 3: Copy service file
echo -e "${BLUE}📄 Installing service file...${NC}"
cp "$SERVICE_FILE" "/etc/systemd/system/"
echo -e "${GREEN}✅ Service file installed${NC}"

# Step 4: Reload systemd and enable service
echo -e "${BLUE}🔄 Reloading systemd daemon...${NC}"
systemctl daemon-reload
echo -e "${GREEN}✅ Systemd daemon reloaded${NC}"

echo -e "${BLUE}🎯 Enabling service...${NC}"
systemctl enable "$SERVICE_NAME"
echo -e "${GREEN}✅ Service enabled for auto-start${NC}"

echo ""
echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
echo ""
echo -e "${YELLOW}📋 Next Steps:${NC}"
echo "1. Copy your WhatsApp Center binary to: $INSTALL_DIR/"
echo "2. Create your .env file at: $INSTALL_DIR/.env"
echo "3. Start the service with: sudo systemctl start $SERVICE_NAME"
echo "4. Check status with: sudo systemctl status $SERVICE_NAME"
echo "5. View logs with: sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo -e "${BLUE}🔧 Service Management Commands:${NC}"
echo "• Start:   sudo systemctl start $SERVICE_NAME"
echo "• Stop:    sudo systemctl stop $SERVICE_NAME"
echo "• Restart: sudo systemctl restart $SERVICE_NAME"
echo "• Status:  sudo systemctl status $SERVICE_NAME"
echo "• Logs:    sudo journalctl -u $SERVICE_NAME -f"
echo "• Disable: sudo systemctl disable $SERVICE_NAME"
echo ""
echo -e "${YELLOW}⚠️  Remember to:${NC}"
echo "• Update environment variables in the service file if needed"
echo "• Configure your .env file with proper OtomaX credentials"
echo "• Ensure the binary has execute permissions"
