# Session History & Current Status

## Accomplishments
- Successfully resolved merge conflicts and merged `task 05` into `main`.
- Restructured the server to use an infinite loop with `listener.Accept()`.
- Implemented a connection cap of 10 clients using a `clients` map (`map[net.Conn]bool`).
- Added concurrency by spinning up a `go handleClient` goroutine for each new connection.
- Used a `sync.Mutex` to prevent data races when modifying the shared `clients` map.
- Handled client disconnects gracefully using a `defer` block to remove them from the map.
- Created an interactive prompt for users to enter their name, with validation to prevent empty (`""`) names.
- Implemented a broadcasting system to send chat messages to everyone *except* the sender.
- Formatted messages to include the timestamp (using Go's reference date `2006-01-02 15:04:05`) and the user's name.
- Added a broadcast announcement when a user joins the chat.

## Current Challenge / Next Step
We are currently trying to add a broadcast announcement when a user *leaves* the chat (`name + " has left our chat..."`). 

This needs to happen inside the `defer` block at the top of `handleClient` so it triggers when the client disconnects. 

**The Problem**: The `defer` block currently cannot access the `name` variable because `name` is declared and assigned using `:=` *below* the `defer` block.

**Next Action**: Split the `name` variable's declaration from its assignment. We need to declare `var name string` *above* the `defer` block, and then just assign it using `name = msgscanner.Text()` *below* the `defer` block.