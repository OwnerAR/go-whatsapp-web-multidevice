# OtomaX API Integration Documentation

This document provides comprehensive documentation for the OtomaX API integration in the WhatsApp Center application.

## Overview

The OtomaX integration allows WhatsApp Center to forward incoming WhatsApp messages to the OtomaX API for processing and receive callbacks with responses. This enables bidirectional communication between WhatsApp and OtomaX systems.

## Security Features

### HMAC-SHA256 Authentication

All requests to OtomaX API use HMAC-SHA256 authentication with the following process:

1. **Metadata Generation**: Create metadata string from request parameters
2. **First HMAC**: Sign metadata with App Key
3. **Second HMAC**: Sign request body with the result from step 2
4. **Final Token**: Combine App ID with the result from step 3

### Payload Expiration (_exp)

All InsertInbox requests include an `_exp` (expiration) field for security:

- **Purpose**: Prevents replay attacks and ensures payload freshness
- **Format**: RFC3339 datetime in UTC+0 (Z timezone)
- **Base Time**: Uses WhatsApp message unix timestamp as base
- **Duration**: 30 seconds from WhatsApp message timestamp
- **Validation**: OtomaX API should reject requests with expired `_exp` timestamps

#### Example _exp Field

```json
{
  "pesan": "Hello World",
  "pengirim": "6285219898489",
  "tipe_pengirim": "W",
  "kode_terminal": 1,
  "_exp": "2025-10-18T20:30:30Z"
}
```

**Note**: The `_exp` timestamp is calculated as: `WhatsApp message timestamp + 30 seconds`. This ensures consistency with the original message timing and prevents replay attacks while allowing sufficient processing time.

## API Endpoints

### InsertInbox

**Endpoint**: `POST /api/insertInbox`

**Purpose**: Forward WhatsApp messages to OtomaX for processing

**Request Structure**:

```json
{
  "pesan": "string",           // Message content from WhatsApp
  "kode_reseller": "string",   // Reseller code (optional)
  "pengirim": "string",        // Sender phone number (without @s.whatsapp.net)
  "tipe_pengirim": "string",   // Sender type (always "W" for WhatsApp)
  "kode_terminal": "integer",  // Terminal code
  "_exp": "datetime"           // Expiration time (UTC+0)
}
```

**Response Structure**:

```json
{
  "ok": true,
  "result": {
    "kode_inbox": 218223,           // Inbox code from OtomaX
    "status": 21,                   // Status code
    "statusDesc": "Sukses Masuk Outbox", // Status description
    "pesan": "Response message"     // Response message (for status 21)
  }
}
```

**Status Codes**:

| Status | Description | Auto Reply |
|--------|-------------|------------|
| 21 | Sukses Masuk Outbox | ✅ Yes (uses `pesan` field) |
| 41 | Bukan Reseller | ✅ Yes (uses `statusDesc`) |
| 42 | Format Salah | ✅ Yes (uses `statusDesc`) |

### SetOutboxCallback

**Endpoint**: `POST /api/setOutboxCallback`

**Purpose**: Set callback URL for receiving OtomaX responses

**Request Structure**:

```json
{
  "url": "https://your-app.com/otomax/callback"
}
```

### GetOutboxCallback

**Endpoint**: `GET /api/getOutboxCallback`

**Purpose**: Retrieve current callback URL configuration

## Message Flow

### 1. WhatsApp Message Received

```
WhatsApp User → WhatsApp Center → Extract Message → Create OtomaX Request
```

### 2. Forward to OtomaX

```
WhatsApp Center → OtomaX InsertInbox API → Process Message → Return Status
```

### 3. Auto Reply (if applicable)

```
OtomaX Response → Check Status (21/41/42) → Send Auto Reply → WhatsApp User
```

### 4. Callback Processing

```
OtomaX System → Callback URL → WhatsApp Center → Process Response
```

## Configuration

### Environment Variables

```bash
# Enable OtomaX integration
OTOMAX_ENABLED=true

# OtomaX API configuration
OTOMAX_API_URL=https://api.otomax.id
OTOMAX_APP_ID=your_app_id
OTOMAX_APP_KEY=your_app_key
OTOMAX_DEV_KEY=your_dev_key

# Default settings
OTOMAX_DEFAULT_RESELLER=default_reseller
OTOMAX_DEFAULT_KODE_TERMINAL=1

# Forwarding options
OTOMAX_FORWARD_INCOMING=true
OTOMAX_FORWARD_OUTGOING=false
OTOMAX_FORWARD_GROUPS=false
OTOMAX_FORWARD_MEDIA=false

# Auto reply settings
OTOMAX_AUTO_REPLY_ENABLED=true
```

### CLI Flags

```bash
# Enable OtomaX
--otomax-enabled=true

# API configuration
--otomax-api-url=https://api.otomax.id
--otomax-app-id=your_app_id
--otomax-app-key=your_app_key
--otomax-dev-key=your_dev_key

# Default settings
--otomax-default-reseller=default_reseller
--otomax-default-kode-terminal=1

# Forwarding options
--otomax-forward-incoming=true
--otomax-forward-outgoing=false
--otomax-forward-groups=false
--otomax-forward-media=false

# Auto reply
--otomax-auto-reply-enabled=true
```

## REST API Endpoints

### WhatsApp Center OtomaX Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/otomax/insert-inbox` | POST | Manual InsertInbox request |
| `/otomax/set-callback` | POST | Set callback URL |
| `/otomax/get-callback` | GET | Get current callback URL |
| `/otomax/test` | POST | Test connection to OtomaX |
| `/otomax/reseller/:code` | GET | Get reseller information |
| `/otomax/balance/:code` | GET | Get reseller balance |
| `/otomax/callback` | POST | Handle OtomaX callbacks |
| `/otomax/health` | GET | Health check |

## Error Handling

### Common Errors

1. **Authentication Failed**
   - Check App ID, App Key, and Dev Key
   - Verify HMAC-SHA256 implementation
   - Ensure proper token generation

2. **Payload Expired**
   - Check system clock synchronization
   - Verify `_exp` field format (RFC3339 UTC+0)
   - Ensure requests are processed within 30 seconds

3. **Invalid Phone Number**
   - Ensure phone number is in international format
   - Remove WhatsApp suffixes (@s.whatsapp.net)
   - Validate phone number format

4. **Network Issues**
   - Check OtomaX API URL accessibility
   - Verify network connectivity
   - Check firewall settings

### Retry Logic

The integration includes automatic retry logic for failed requests:

- **Max Attempts**: 3 retries
- **Backoff**: Exponential (1s, 2s, 4s)
- **Timeout**: 30 seconds per attempt

## Best Practices

### Security

1. **Use HTTPS**: Always use HTTPS for OtomaX API URLs
2. **Secure Keys**: Store API keys securely (environment variables, not in code)
3. **Validate Callbacks**: Always verify callback authenticity
4. **Monitor Expiration**: Ensure `_exp` timestamps are properly set

### Performance

1. **Async Processing**: Use goroutines for non-blocking operations
2. **Connection Pooling**: Reuse HTTP connections when possible
3. **Rate Limiting**: Respect OtomaX API rate limits
4. **Error Logging**: Log all errors for debugging

### Reliability

1. **Idempotency**: Handle duplicate requests gracefully
2. **Timeout Handling**: Set appropriate timeouts for API calls
3. **Graceful Degradation**: Continue operation if OtomaX is unavailable
4. **Health Monitoring**: Monitor OtomaX API health

## Troubleshooting

### Debug Logging

Enable debug logging to troubleshoot issues:

```bash
./whatsapp rest --debug=true --otomax-enabled=true
```

### Common Issues

1. **Messages not forwarded to OtomaX**:
   - Check `OTOMAX_ENABLED` setting
   - Verify forwarding options
   - Check OtomaX API connectivity

2. **Auto replies not sent**:
   - Check `OTOMAX_AUTO_REPLY_ENABLED` setting
   - Verify status codes (21, 41, 42)
   - Check WhatsApp client connectivity

3. **Authentication errors**:
   - Verify API credentials
   - Check HMAC implementation
   - Ensure proper token generation

### Log Examples

**Successful InsertInbox Request**:
```
INFO[2025-10-18T20:30:00Z] Forwarding WhatsApp message to OtomaX InsertInbox: sender=6285219898489@s.whatsapp.net, message=Hello World
INFO[2025-10-18T20:30:01Z] OtomaX InsertInbox response: &{Ok:true Result:{KodeInbox:218223 Status:21 StatusDesc:Sukses Masuk Outbox Pesan:Message processed successfully}}
INFO[2025-10-18T20:30:01Z] OtomaX auto reply sent successfully to 6285219898489@s.whatsapp.net: Message processed successfully
```

**Failed Authentication**:
```
ERROR[2025-10-18T20:30:00Z] Failed to send request to OtomaX: authentication failed
ERROR[2025-10-18T20:30:00Z] Failed to send request to OtomaX: failed to send request to OtomaX: authentication failed
```

## Integration Examples

### Node.js Webhook Handler

```javascript
const express = require('express');
const app = express();

app.use(express.json());

// Handle OtomaX callbacks
app.post('/otomax/callback', (req, res) => {
    const { kode, status, message, pesan, pengirim } = req.body;
    
    console.log('Received OtomaX callback:', {
        kode,
        status,
        message,
        pesan,
        pengirim
    });
    
    // Process callback based on status
    switch (status) {
        case 21:
            console.log('Success:', pesan);
            break;
        case 41:
            console.log('Not a reseller:', message);
            break;
        case 42:
            console.log('Wrong format:', message);
            break;
        default:
            console.log('Unknown status:', status, message);
    }
    
    res.status(200).json({ status: 'success' });
});

app.listen(3000, () => {
    console.log('OtomaX callback handler listening on port 3000');
});
```

### Python Webhook Handler

```python
from flask import Flask, request, jsonify
import logging

app = Flask(__name__)
logging.basicConfig(level=logging.INFO)

@app.route('/otomax/callback', methods=['POST'])
def handle_otomax_callback():
    data = request.get_json()
    
    kode = data.get('kode')
    status = data.get('status')
    message = data.get('message')
    pesan = data.get('pesan')
    pengirim = data.get('pengirim')
    
    app.logger.info(f'Received OtomaX callback: {data}')
    
    # Process callback based on status
    if status == 21:
        app.logger.info(f'Success: {pesan}')
    elif status == 41:
        app.logger.info(f'Not a reseller: {message}')
    elif status == 42:
        app.logger.info(f'Wrong format: {message}')
    else:
        app.logger.info(f'Unknown status: {status} - {message}')
    
    return jsonify({'status': 'success'})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=3000)
```

## Monitoring and Metrics

### Key Metrics to Monitor

1. **Request Success Rate**: Percentage of successful OtomaX API calls
2. **Response Time**: Average time for OtomaX API responses
3. **Auto Reply Rate**: Percentage of messages that trigger auto replies
4. **Error Rate**: Frequency of authentication and network errors
5. **Callback Processing**: Success rate of callback handling

### Health Checks

The integration provides health check endpoints:

```bash
# Check OtomaX integration health
curl http://localhost:3000/otomax/health

# Test OtomaX API connection
curl -X POST http://localhost:3000/otomax/test \
  -H "Content-Type: application/json" \
  -d '{"phone": "6285219898489"}'
```

## Changelog

### v7.10.0
- Added `_exp` (expiration) field to InsertInbox requests
- Improved security with payload expiration
- Enhanced error handling and logging
- Added comprehensive documentation

### v7.9.0
- Initial OtomaX API integration
- HMAC-SHA256 authentication
- Auto-reply functionality
- Callback handling system
