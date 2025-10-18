#!/bin/bash

# Debug WhatsApp Center Service
# Script untuk debugging systemd service

echo "🔍 WhatsApp Center Service Debug Information"
echo "=========================================="

echo ""
echo "📁 Current Directory and Permissions:"
echo "------------------------------------"
pwd
ls -la

echo ""
echo "📋 Service Status:"
echo "----------------"
sudo systemctl status whatsapp-center --no-pager -l

echo ""
echo "📄 Service Configuration:"
echo "-----------------------"
sudo systemctl cat whatsapp-center

echo ""
echo "📊 Recent Logs:"
echo "--------------"
sudo journalctl -u whatsapp-center --no-pager -l -n 20

echo ""
echo "🔧 System Information:"
echo "-------------------"
echo "OS: $(cat /etc/os-release | grep PRETTY_NAME)"
echo "Kernel: $(uname -r)"
echo "Systemd Version: $(systemctl --version | head -1)"

echo ""
echo "👤 User Information:"
echo "------------------"
echo "Current User: $(whoami)"
echo "Service User: $(grep User= /etc/systemd/system/whatsapp-center.service | cut -d'=' -f2)"
echo "Service Group: $(grep Group= /etc/systemd/system/whatsapp-center.service | cut -d'=' -f2)"

echo ""
echo "📂 Directory Permissions:"
echo "-----------------------"
echo "Working Directory: $(grep WorkingDirectory= /etc/systemd/system/whatsapp-center.service | cut -d'=' -f2)"
if [ -d "$(grep WorkingDirectory= /etc/systemd/system/whatsapp-center.service | cut -d'=' -f2)" ]; then
    ls -la "$(grep WorkingDirectory= /etc/systemd/system/whatsapp-center.service | cut -d'=' -f2)"
else
    echo "❌ Working directory does not exist!"
fi

echo ""
echo "🔍 Binary Information:"
echo "--------------------"
BINARY_PATH=$(grep ExecStart= /etc/systemd/system/whatsapp-center.service | cut -d'=' -f2 | cut -d' ' -f1)
echo "Binary Path: $BINARY_PATH"
if [ -f "$BINARY_PATH" ]; then
    ls -la "$BINARY_PATH"
    file "$BINARY_PATH"
else
    echo "❌ Binary does not exist at: $BINARY_PATH"
fi

echo ""
echo "🌐 Network Information:"
echo "--------------------"
echo "Listening Ports:"
sudo netstat -tlnp | grep -E ":(3000|8080|80|443)"

echo ""
echo "💾 Memory and CPU:"
echo "----------------"
free -h
df -h /
