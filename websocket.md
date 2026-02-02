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
Sent periodically or on change. The first argument is a JSON stringified object.

```json
{
  "event": "stats",
  "args": ["{\"memory_bytes\":11419623424,\"cpu_absolute\":18.8,\"network\":{\"rx_bytes\":3682862139,\"tx_bytes\":102412526},\"uptime\":11179,\"disk_bytes\":200050167808,\"audio\":[{\"id\":\"player1\",\"name\":\"Chrome\",\"playing\":true,\"artist\":\"Hoobastank\",\"title\":\"The Reason\",\"album\":\"Fallen\",\"art_url\":\"/v1/img/tmp/L3RtcC8ub3JnLmNocm9taXVtLkNocm9taXVtLk5Vbnl0bQ==\",\"timestamp\":115,\"duration\":232},{\"id\":\"player2\",\"name\":\"Firefox\",\"playing\":false,\"artist\":\"Artist Name\",\"title\":\"Video Title\",\"timestamp\":1671,\"duration\":3215}],\"wifi\":{\"ssid\":\"Bazinga! 5G\",\"connected\":true},\"battery\":{\"percentage\":100,\"plugged_in\":true},\"volume\":88,\"backlight\":64}"]
}
```

*Note: The `art_url`field contains a api endpoint to fetch the album art image. The path is base64 URL encoded.*

### `session expiring`
Sent 4 minutes before disconnection.
```json
{
  "event": "session expiring ",
  "args": ["[10:30:45]: Your Session will expire"]
}
```

## Outgoing Events (Client to Server)

### Audio Control
Control media playback by targeting a specific player ID received in the `stats` event (e.g., "player1").

| Event | Argument | Description |
|-------|----------|-------------|
| `audio-play-pause` | `"playerID"` | Toggle play/pause for the specific player |
| `audio-next` | `"playerID"` | Skip to next track |
| `audio-previous` | `"playerID"` | Go to previous track |
| `audio-stop` | `"playerID"` | Stop playback |

**Example:**
```json
{
  "event": "audio-play-pause",
  "args": ["player1"]
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
