# 5) Broadcast chat messages to everyone
## Goal

Anything a client sends should be delivered to all other connected clients.

## Steps

1. Each client handler reads lines in a loop.
2. Send messages into a central `broadcast` channel.
3. One goroutine reads from `broadcast` and writes to all clients.
4. Protect the clients list with a mutex OR do all list changes via channels.

## Done when

- Messages from one client appear on all others.
- No data races / panics.