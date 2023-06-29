# Cue

A real-time location service for identifying nearby users.

## Commands

Start all microservices.

```bash
./scripts/start-docker.sh
```

Start the gateway service.

```bash
./scripts/start-gateway.sh
```

Start the proximity service.

```bash
./scripts/start-proximity.sh
```

## Instructions

1. Create a new `.env` file in the root directory with the following variables:

```
REDIS_URL=redis://redis:6379
REDIS_GATEWAY_CHANNEL_IN=gateway.messages_in
REDIS_PROXIMITY_CHANNEL_IN=proximity.messages_in

AUTH0_DOMAIN=YOUR_AUTH0_DOMAIN
AUTH0_CLIENT_ID=YOUR_AUTH0_CLIENT_ID
AUTH0_CLIENT_SECRET=YOUR_AUTH0_CLIENT_SECRET
AUTH0_CALLBACK_URL=YOUR_AUTH0_CALLBACK_URL

PORT=8080
```

2. Start the application:

```bash
./scripts/start-docker.sh
```

3. Navigate to the following link, authenticate, then copy the `session-cookie`.

4. Navigate to the following [link](https://www.piesocket.com/websocket-tester?url=ws://localhost:8080/ws), `Connect`, and start sending messages e.g.

```
{ "sessionId": "session-cookie", "eventType": 1, "body": "{ \"user\": \"JohnDoe\", \"lat\": 37.7749, \"long\": -122.4194, \"timestamp\": \"2023-06-27T10:30:00Z\" }" }
```
