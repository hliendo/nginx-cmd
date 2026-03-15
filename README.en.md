# Nginx Command Tool (nginx-cmd)
This project does not aim to replace the brilliance of **Caddy** or the convenience of **Nginx Proxy Manager**. Both are remarkable achievements of the open source community, and their creators put tremendous effort into building them.

Instead, `nginx-cmd` positions itself as a **minimalist alternative** for sysadmins who need:
- Absolute control over Nginx configurations
- Modularity through `include` directives
- A clean lifecycle with no technical debt (always destroy and recreate, never edit)

Where Caddy shines with automation and simplicity, and Nginx Proxy Manager provides a user-friendly interface backed by a database, `nginx-cmd` focuses on **raw Nginx power combined with unattended certificate management** — without adding extra layers or expanding the attack surface.

In short:  
`nginx-cmd` is not a competitor, but a complementary tool for those who value **clarity, idempotence, and structural control** in their infrastructure.

**`nginx-cmd`** is not an AI-driven template engine, nor an Nginx-Proxy-Manager clone full of web dependencies, nor does it intend to replace the power of human handwriting in a `.conf` file.

It is an agile, strict, and immutable orchestrator, written in **Golang**, born within the CDC architecture to eliminate the fragility of historical Bash scripts. It manages the secure insertion, initialization, and vital lifecycle (ACME Bootstrap and trust stores) of domains, **without yielding a single comma of structural control**.

## 🏗️ History and the "Industrial Truth"

Originally, we managed raw Nginx by installing Certbot and handling dependencies manually (or relying on fragile bash scripts).  
We later experimented with **Caddy**: brilliant and easy to use, but in complex architectures we found ourselves trading some granular control for simplicity (e.g., custom modules, heavy gRPC traffic, strict caching).  
We also tried **Nginx Proxy Manager**: convenient with its visual layer and database, but this added complexity and a broader attack surface, moving away from the "truth" of pure `.conf` files.  

From these experiences, this project was born: **the union of Caddy’s unattended magic with the untouchable, raw power of Nginx**.

## ⚖️ Architectural Doctrine of `nginx-cmd`

To scale this project to the community, two unbreakable golden rules apply:

### 1. Business Configuration is Human ("Features as Includes")
`nginx-cmd` generates idiot-proof Server Blocks ("Stubs"). All the real operational intelligence (Buffer sizes, WebSocket upgrades, proxy timeouts, CORS) is not injected via CLI flags, nor processed by Go. **That logic is retained by the Sysadmin on their local disk**, through modular files mounted in `conf.d/domains/includes/` (e.g., `proxy-websocket.conf`). The binary only injects the `include` instruction pointing to them. This guarantees total portability and zero need to recompile the binary for every infrastructure tweak.

### 2. Idempotency and Destruction: The "No Edit" Rule
`nginx-cmd` **does not have** an `edit` command.
Modifying partialized domains via CLI implies bi-directional parsing and managing state variables that corrupt easily (e.g., if writing fails halfway, Nginx dies).
If you made a mistake when creating a domain (e.g., wrong target or typed SPA instead of Proxy):
```bash
# Correct workflow:
nginx-cmd remove my-web.com
nginx-cmd add my-web.com proxy --target 10.0.0.5:80
```
**Destruction and Recreation.** Zero debt, state always clean.

## 🚀 Installation & Usage (Docker First)

`nginx-cmd` is designed specifically to run inside its Docker container (Alpine with Brotli module and ACME lifecycle).

### 1. Deploy with Docker Compose
Clone the repository, configure your `.env` (using `.env.example` as a template), and start the services:
```bash
docker compose up -d --build
```

### 2. Run Commands from the Host
To manage domains without manually entering the container (`docker exec`), use the included wrapper script at the root:

```bash
# Add execution permissions (if needed)
chmod +x nginx-cmd.sh

# List active domains
./nginx-cmd.sh list

# Add a domain (Reverse Proxy)
./nginx-cmd.sh add mydomain.com proxy --target 10.0.0.5:8080
```

*Find a detailed breakdown of routing behaviors in [USAGE.en.md](USAGE.en.md).*

