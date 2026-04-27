#
# TCP Server Commands

## 1. TCP Server (listener)

This is how you start your chat server:

    ln, err := net.Listen("tcp", ":8989")

* "tcp" → protocol
* ":8989" → port
* ln → listener (waits for clients)


## 2. Accept connections

    conn, err := ln.Accept()

* Blocks until a client connects
* Returns a net.Conn (connection object)


## 3. Connection (net.Conn)

This is VERY important.
Think of it like a pipe between client and server.

    conn.Write([]byte("hello"))
    conn.Read(buffer)

* Write → send data
* Read → receive data

## 4. Client side (connect to server)

    conn, err := net.Dial("tcp", "localhost:8989")

* This is how a client connects

---
---
### Real mental model
* Listener = door 🚪
* Accept() = someone knocking
* Conn = private phone line 📞

Each client gets:
* 👉 their own Conn
* 👉 handled in a goroutine

---
---
