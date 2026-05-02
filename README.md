# TCP-Chat

A concurrent TCP chat server written in Go. This project allows multiple clients to connect, choose a username, and broadcast messages to everyone else in the chat room.

## Features

- **Concurrent Connections**: Handles up to 10 clients simultaneously using goroutines.
- **Unique Usernames**: Each client must provide a unique, non-empty username upon connecting.
- **Message Broadcasting**: Messages sent by one client are broadcast to all other connected clients.
- **Graceful Disconnects**: The server properly handles clients leaving the chat.
- **Customizable Port**: The server can be run on a custom port provided as a command-line argument.

## How to Run

1.  **Clone the repository:**
    ```sh
    git clone https://platform.zone01.gr/git/gelafros/net-cat
    cd net-cat
    ```

2.  **Run the server:**

    *   To run on the default port (8989):
        ```sh
        go run .
        ```

    *   To specify a custom port:
        ```sh
        go run . 9000
        ```

## How to Connect

You can connect to the chat server using a tool like `netcat` or `telnet`.

```sh
nc localhost 8989
```

Once connected, you will be prompted to enter a name. After providing a unique name, you can start sending messages to the chat room.
