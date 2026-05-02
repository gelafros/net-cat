# Project History & Changelog

## Phase 1: Server Foundation
- **Initialization**: Built standard TCP server running on default port `8989` or user-defined port with `[USAGE]` error handling.
- **Connection Handshake**: Implemented the NetCat Linux logo welcome message and username prompt.
- **Validation**: Added checks to require non-empty names and prevent duplicate names from joining.
- **Capacity Control**: Implemented a maximum connection limit of 10 concurrent clients.
- **Concurrency**: Added `sync.Mutex` locks to protect the client map and history array from data races.

## Phase 2: Core Chat Features
- **Formatted Messages**: Implemented standard message formatting: `[timestamp][name]: message`.
- **Chat History Sync**: Implemented a global history array that uploads all previous chat context to newly joined users.
- **System Announcements**: Added global broadcast notifications when a user joins or leaves the chat.
- **Empty Message Filtering**: Prevented the server from broadcasting empty or whitespace-only messages.

## Phase 3: Scalability & Architecture
- **Client-Writer Pattern**: Replaced blocking network writes with dedicated buffered channels (`chan string`, size 100) and personal background goroutines for each client.
- **Non-Blocking Broadcasts**: Updated the global Broadcaster to use non-blocking `select` statements to prevent slow clients from freezing the server.
- **Project Restructure**: Moved core business logic out of `TCPChat.go` and `handlers/` into a standard Go `internal/` package structure.

## Phase 4: Bonus Features & User Experience
- **Dynamic Usernames**: Added the `--name` command to allow users to change their display names mid-chat, with duplicate checking.
- **Chat Colors**: Added the `--color` command so users can customize their message colors (red, green, yellow, blue, magenta, cyan).
- **Help Menu**: Added `/help` to display available commands.
- **Fun Commands**: Added `--tableflip` for ASCII art messaging.

## Phase 5: Polish & Security (Latest)
- **Server Logging**: Added `io.MultiWriter` logging to both the terminal and `server_logs.txt` (ignored in `.gitignore`), standardizing all log formats.
- **Private Messaging**: Implemented the `--whisper` command for direct, private communication between clients.
- **Security Enhancements**: Added a robust ANSI escape sequence stripper to sanitize user input and prevent terminal UI injection (e.g., arrow key overwrites, backspaces).
- **Command Refactor**: Centralized all command parsing (`/help`, `--name`, `--color`, `--tableflip`, `--whisper`) into `internal/commands.go` using the Separation of Concerns pattern.
- **Branch Merge**: Merged the feature branch `elafrostoxanakanw` into `main` and safely removed the deprecated `handlers/` directory.