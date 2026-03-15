# Usage Manual: nginx-cmd (Domain Tree)

`nginx-cmd` is the strict orchestrator of the CDC architecture for Nginx. Its sole purpose is to inject configurations based on **two unique network paradigms**: **SPA** (Static Frontend) and **Proxy** (Backend Intermediary).

When executing `nginx-cmd add <domain> <type>`, the binary injects a pre-defined tree of headers ('Includes') into the server.

---

## 🌳 Structural Tree of nginx-cmd

```text
nginx-cmd
│
├── 1. SPA (Single Page Application)
│   └── Command: nginx-cmd add domain.com spa
│   └── (No Additional Switches Available)
│
└── 2. Proxy (Reverse Proxy)
    ├── Command: nginx-cmd add domain.com proxy --target <IP:PORT>
    │
    ├── [+ Optional Switch] --ws     (WebSocket Protocol Support)
    ├── [+ Optional Switch] --grpc   (gRPC Protocol Support)
    └── [+ Optional Switch] --stream (Buffering-Off for SSE/Video streams)
```

---

## 1. Site Type: SPA (Single Page Application)
**Typical Use Case:** Static websites developed in React, Vue, Svelte, or plain HTML/JS.
**Base Command:** `nginx-cmd add <domain.com> spa`

An SPA requires a classic web server behavior. Nginx acts as a "File Server", reading `.html` files straight from the container's hard drive (`/opt/nginx/html/<domain>`) and implementing the necessary routing logic for JS frameworks (`try_files $uri /index.html;`).

### ⚙️ Available Switches (SPA)
**NONE.**
- **Why?** Because in an SPA, client-server communication occurs exclusively via HTTP GET requests for static files on the local disk. Nginx *does not forward* traffic to a third party. Therefore, tunneling switches like WebSocket, gRPC, or Chunking (Streaming) have no logical entity because there is no active "Backend" connected to Nginx's root `location /`.

---

## 2. Site Type: Proxy (Reverse Proxy)
**Typical Use Case:** APIs in Node.js, Python, Rust, Go, exposed databases, or services like Portainer, Gitea, etc.
**Base Command:** `nginx-cmd add <domain.com> proxy --target <IP:PORT>`

A Proxy requires Nginx to act as a blind, delegating "Tunnel". Nginx receives traffic from the internet and re-assembles it towards an internal IP in our network.

### ⚙️ Mandatory Switches
- `--target <HOST:PORT>` (e.g., `--target 10.10.10.200:8080`)
  - **Function:** Tells Nginx exactly to which IP and port to send the re-assembled traffic.

### ⚙️ Optional Switches (Features injected as Includes)

The Proxy type admits three (3) behavioral modifiers (switches) to adapt Nginx to the technical requirements of the connected backend.

#### A. Switch `--ws` (WebSocket Support)
- **Function:** Injects the `proxy-websocket.conf` include file into Nginx.
- **Technical Effect:** Forces Nginx to intercept the client's HTTP Request, look for the `Upgrade: websocket` header, and transform the standard "Ask-Reply (HTTP)" request into a **"Long-Lived Bidirectional Pipe"**.
- **Why?** Because if this "Upgrade Header" is not proxied, the Nginx server severs the tunnel, assuming the initial HTTP download finished. Fundamental for chats or real-time P2P synchronization.

#### B. Switch `--grpc` (gRPC Protocol Support)
- **Function:** Activates gRPC routing and proxy-pass (`proxy-grpc.conf`).
- **Technical Effect:** Replaces the traditional `http://` tunnel with `grpc://`. Applies the strict HTTP/2 support required by gRPC and omits obsolete proxy headers that would corrupt Google's Protobuf packets (used in high-sync microservices).
- **Why?** Because a backend written in gRPC does not understand generic generic HTTP; it reads direct packaged binary. Nginx needs a surgical command to stop acting as a Web server and become a blind HTTP/2 gRPC orchestrator.

#### C. Switch `--stream` (Server-Sent Events / Chunking)
- **Function:** Disables network buffering (`proxy-stream.conf`).
- **Technical Effect:** Instructs Nginx (`proxy_buffering off;`) not to attempt to save the backend's response in its RAM until fully packaged. Every byte Nginx receives from the Backend is instantly pushed to the final user's browser. It also extends connection Timeouts to the maximum (3600s).
- **Why?** Nginx is "Anxious" by default: if you request a 5GB movie, it will try to download the entire 5GB into its memory **before sending it to the user**. With this switch on, if you have APIs yielding events drop by drop (SSE, ChatGPT replies, Video Streaming/IP-Cams), Nginx disables the central network brake and becomes transparent.
