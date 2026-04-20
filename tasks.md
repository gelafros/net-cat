# 3) Welcome screen + ask name (non-empty)
## Goal

On connect, show the ASCII welcome and ask for a non-empty name.

## Steps

1. Write the welcome banner + linux logo to the connection.
2. Prompt exactly: `[ENTER YOUR NAME]:`
3. Read a line using bufio.Reader.
4. If name is empty/only spaces → prompt again.
5. Save the final name in the client struct.

## Done when

- Client cannot enter empty name.
- Prompt repeats until valid.