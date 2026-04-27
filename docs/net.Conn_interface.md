#
# net.Conn Interface

    type Conn interface {
    Read(b []byte) (n int, err error)
    Write(b []byte) (n int, err error)
    Close() error

    LocalAddr() Addr
    RemoteAddr() Addr

    SetDeadline(t time.Time) error
    SetReadDeadline(t time.Time) error
    SetWriteDeadline(t time.Time) error
    }

So when you do:

    conn, _ := ln.Accept()

👉 conn is an interface value

---
## What you can do with Conn

### 1. 📥 Read data (client → server)

    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)

- n → how many bytes you got
- buffer[:n] → actual message

---

### 2. 📤 Send data (server → client)

    conn.Write([]byte("Hello client\n"))

---

### 3. ❌ Close connection

    conn.Close()

- Important when client leaves.

---

### 4. 🌍 Get client info
    addr := conn.RemoteAddr()
    fmt.Println(addr.String())

Example output:

    127.0.0.1:54321

👉 This is how you identify clients (IP + port)

---

### 5. 🖥️ Get server info

    conn.LocalAddr()

Usually:

    localhost:8989
---

### 6. ⏱️ Set timeouts (advanced but useful)

    conn.SetReadDeadline(time.Now().Add(10 * time.Second))

👉 Prevents hanging connections

---

## Summary

- Read() → get messages
- Write() → send messages
- Close() → disconnect
- RemoteAddr() → identify client

### Critical Detail
---
Read() is blocking.

That means:

    conn.Read(buffer)

👉 waits until:

- client sends something OR
- connection closes

That’s why you MUST use:

    go handleClient(conn)

Otherwise your server = dead after 1 client.    