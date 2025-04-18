---
title: "Building a Simple TCP Client/Server in Zig: Complete Breakdown"
date: 2025-04-18
excerpt: "A comprehensive guide to socket programming in Zig, featuring a complete breakdown of TCP client and server implementation"
---

## Building a Simple TCP Client/Server in Zig: Complete Breakdown

### Introduction

Socket programming is an essential skill for developers working on networked applications. Zig, as a modern systems programming language, offers powerful yet straightforward tools for implementing network communication. In this article, we'll explore how to build a simple TCP client-server application in Zig, breaking down the code for clarity and understanding.

### Project Structure

Our TCP socket programming project consists of two main files:

```
zig-sockets/
│
├── src/
│   ├── client.zig      // Client implementation
│   └── server.zig      // Server implementation
│
└── build.zig           // Build script (not shown in this example)
```

Let's examine each file in detail to understand how they work together to create a functional client-server system.

### `server.zig` - TCP Echo Server

The server component listens for incoming connections, accepts them, and echoes back any data received from clients. Here's the complete implementation:

```zig
const std = @import("std");
const posix = std.posix;

pub fn main() !void {
    const allocator = std.heap.page_allocator;

    const address = try std.net.Address.parseIp("127.0.0.1", 9000);
    const sockfd = try posix.socket(posix.AF.INET, posix.SOCK.STREAM, 0);
    try posix.bind(sockfd, &address.any, address.getOsSockLen());
    try posix.listen(sockfd, 10);

    while (true) {
        var client_addr: std.net.Address = undefined;
        var addr_len: u32 = @sizeOf(std.net.Address);
        const client_fd = try posix.accept(sockfd, &client_addr.any, &addr_len, 0);

        var buf: [1024]u8 = undefined;
        const n = try posix.read(client_fd, buf[0..]);
        _ = try posix.write(client_fd, buf[0..n]);
        _ = posix.close(client_fd);
    }
}
```

Let's break down the key components of the server:

1. **Socket Initialization**:
   ```zig
   const address = try std.net.Address.parseIp("127.0.0.1", 9000);
   const sockfd = try posix.socket(posix.AF.INET, posix.SOCK.STREAM, 0);
   ```
   This creates a TCP socket (`SOCK.STREAM`) in the IPv4 domain (`AF.INET`) and sets up an address structure for localhost on port 9000.

2. **Binding and Listening**:
   ```zig
   try posix.bind(sockfd, &address.any, address.getOsSockLen());
   try posix.listen(sockfd, 10);
   ```
   The server binds the socket to the specified address and begins listening for connections, with a backlog queue of 10 pending connections.

3. **Main Server Loop**:
   ```zig
   while (true) {
       // Accept and handle connections
   }
   ```
   The server continuously accepts new connections and processes them.

4. **Connection Handling**:
   ```zig
   const client_fd = try posix.accept(sockfd, &client_addr.any, &addr_len, 0);
   var buf: [1024]u8 = undefined;
   const n = try posix.read(client_fd, buf[0..]);
   _ = try posix.write(client_fd, buf[0..n]);
   _ = posix.close(client_fd);
   ```
   For each client connection, the server reads data into a buffer, writes it back to the client (echo), and then closes the connection.

### `client.zig` - TCP Client

The client connects to the server, sends user input, and displays the server's response. Here's the complete implementation:

```zig
const std = @import("std");
const posix = std.posix;
pub fn main() !void {
    const allocator = std.heap.page_allocator;
    const address = try std.net.Address.parseIp("127.0.0.1", 9000);
    const sockfd = try posix.socket(posix.AF.INET, posix.SOCK.STREAM, 0);
    try posix.connect(sockfd, &address.any, address.getOsSockLen());
    var buf: [1024]u8 = undefined;
    const stdin = std.io.getStdIn().reader();
    const stdout = std.io.getStdOut().writer();
    try stdout.print("💬 Connected! Type a line and press Enter:\n> ", .{});
    const len = try stdin.readUntilDelimiterOrEof(buf[0..], '\n');
    _ = try posix.write(sockfd, buf[0..len.?]);
    const n = try posix.read(sockfd, buf[0..]);
    try stdout.print("🔁 Echoed back: {s}\n", .{buf[0..n]});
    _ = posix.close(sockfd);
}
```

Let's break down the key components of the client:

1. **Socket Creation and Connection**:
   ```zig
   const address = try std.net.Address.parseIp("127.0.0.1", 9000);
   const sockfd = try posix.socket(posix.AF.INET, posix.SOCK.STREAM, 0);
   try posix.connect(sockfd, &address.any, address.getOsSockLen());
   ```
   The client creates a socket and connects to the server at localhost:9000.

2. **Input/Output Setup**:
   ```zig
   const stdin = std.io.getStdIn().reader();
   const stdout = std.io.getStdOut().writer();
   ```
   Sets up standard input and output streams for user interaction.

3. **User Input**:
   ```zig
   try stdout.print("💬 Connected! Type a line and press Enter:\n> ", .{});
   const len = try stdin.readUntilDelimiterOrEof(buf[0..], '\n');
   ```
   Prompts the user for input and reads a line of text.

4. **Data Exchange**:
   ```zig
   _ = try posix.write(sockfd, buf[0..len.?]);
   const n = try posix.read(sockfd, buf[0..]);
   ```
   Sends the user's input to the server and receives the echoed response.

5. **Result Display**:
   ```zig
   try stdout.print("🔁 Echoed back: {s}\n", .{buf[0..n]});
   ```
   Displays the server's response to the user.

### Error Handling in Zig

One of Zig's standout features is its approach to error handling. Throughout both files, you'll notice extensive use of the `try` keyword:

```zig
const sockfd = try posix.socket(posix.AF.INET, posix.SOCK.STREAM, 0);
```

This syntax performs two critical functions:
1. It unwraps the result of the function if successful
2. If an error occurs, it returns the error up the call stack

This approach makes error handling explicit and concise. The `!void` return type on the `main()` function indicates that it can return errors, which the Zig runtime will handle appropriately.

### Building and Running the Application

To build and run this application, you'll need to:

1. Install Zig (if not already installed)
2. Save the code into respective files (`server.zig` and `client.zig`)
3. Build the executables:
   ```bash
   zig build-exe server.zig
   zig build-exe client.zig
   ```
4. Run the server in one terminal window:
   ```bash
   ./server
   ```
5. Run the client in another terminal window:
   ```bash
   ./client
   ```

The client will prompt you to type a message, which it will send to the server. The server will echo it back, and the client will display the response.

### Key Observations About Zig for Network Programming

1. **Explicit Error Handling**: Zig forces developers to address errors, leading to more robust code.

2. **Memory Safety**: Zig's control over memory allocation helps prevent common bugs in network code.

3. **Low-Level Access**: Direct POSIX binding provides fine-grained control over socket operations.

4. **Readability**: Despite being a systems language, Zig code remains relatively readable and concise.

### Potential Enhancements

This basic implementation can be extended in several ways:

1. **Concurrent Connections**: Implement threading or async I/O to handle multiple clients simultaneously.

2. **Persistent Connections**: Modify the server to keep client connections open for multiple exchanges.

3. **Protocol Implementation**: Add a simple application protocol on top of TCP.

4. **Error Recovery**: Implement more sophisticated error handling and recovery mechanisms.

5. **Configuration Options**: Add command-line arguments for IP address, port, and other settings.

### Conclusion

This example demonstrates Zig's capability as a systems programming language for network applications. Its approach to error handling, memory management, and direct system access makes it well-suited for building reliable networked services.

By breaking down the code into clear, focused files, we've created a maintainable structure that can be extended for more complex networking applications. Whether you're just starting with Zig or looking to enhance your network programming skills, this simple TCP client-server implementation provides a solid foundation to build upon.

---
