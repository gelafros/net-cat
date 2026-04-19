# 2) Accept connections + limit to 10

## Goal

Accept multiple clients but cap at 10 active connections.

## Steps

1. Keep a counter / list of active clients.
2. When a connection arrives:
    - If active >= 10: send a short message like `Server busy` and close.
    - Else accept and start a goroutine to handle that client.
3. When a client disconnects, decrement/remove from the active list.

## Done when

- 11th client gets refused.
- Existing clients stay connected.