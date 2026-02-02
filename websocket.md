# WebSocket Documentation

## Overview
Secure WebSocket connection for monitoring system resources. The connection has a hard expiration of 20 minutes.

## Connection Flow

1. **Obtain Access Token**
   - `POST /v1/login` with credentials
   - Response includes JWT login token

2. **Obtain WebSocket Details**
   - `GET /v1/websocket` with format `Authorization: Bearer [LOGIN_TOKEN]`
   - Response:
     ```json
     {
       "object": "websocket_token",
       "data": {
         "token": "eyJ...",
         "socket": "wss://host:port/v1/monitor/[uuid]/ws"
       }
     }
     ```

3. **Connect**
   - Connect to the returned `socket` URL.
   - **Immediately** send authentication frame:
     ```json
     {
       "event": "auth",
       "args": ["WEBSOCKET_TOKEN"]
     }
     ```

## Incoming Events

### `stats`
Sent periodically or on change.
```json
{
  "event": "stats",
  "args": ["{\"memory_bytes\":123...}"]
}
```
*Note: The first argument is a JSON stringified object containing: `memory_bytes`, `cpu_absolute`, `network`, `uptime`, `disk_bytes`, `audio` (metadata), `wifi`, `battery`, etc.*

### `session expiring`
Sent 4 minutes before disconnection.
```json
{
  "event": "session expiring ",
  "args": ["[10:30:45]: Your Session will expire"]
}
```

## Close Codes

| Code | Description | Action |
|------|-------------|--------|
| `1000` | Normal Closure | None |
| `1001` | Going Away | None |
| `1006` | Abnormal | Attempt Reconnect (5s) |
| `4001` | Auth Failed | Login Again |
| `4004` | Token Expired | Refresh Token and Reconnect |
