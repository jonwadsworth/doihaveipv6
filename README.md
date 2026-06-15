# doihaveipv6.com

A simple, honest IPv6 detection service. Tells you whether you're connecting to the internet over IPv6 or IPv4 — with a deliberately obnoxious YES or NO.

**[doihaveipv6.com](https://doihaveipv6.com)**

## How it works

Most "what's my IP" services just show whatever address your connection arrived from. That's fine, but it only answers half the question — if you happened to connect over IPv4 even though your network supports IPv6, you'd get a false NO.

This service uses a **dual-probe technique** to give an accurate answer:

1. The page's JavaScript fires two parallel requests:
   - `ipv4.doihaveipv6.com/api` — has an **A record only**, so only IPv4 clients can reach it
   - `ipv6.doihaveipv6.com/api` — has an **AAAA record only**, so only IPv6 clients can reach it
2. Whichever probe succeeds tells us which protocols actually work for you, regardless of which one you used to load the main page
3. Falls back to server-side IP detection if the probe subdomains aren't reachable

The DNS split is the trick. By restricting each subdomain to a single record type, the network itself does the filtering.

## Architecture

```
Browser
  │
  ├─── doihaveipv6.com (A + AAAA) ──────────────────────────────┐
  ├─── ipv4.doihaveipv6.com (A only)  · JS probe ──────────────┤
  └─── ipv6.doihaveipv6.com (AAAA only) · JS probe ────────────┤
                                                                  ▼
                                              nginx (SNI + virtual hosting)
                                                                  │
                                              Go binary · port 3000
                                              ├── GET /        HTML page
                                              ├── GET /api     IP detection (JSON)
                                              └── GET /geo     ISP lookup → ipwho.is
```

## Stack

- **Go** (stdlib only, no framework) — single binary, ~8MB, ~1.2MB RAM idle
- **nginx** — reverse proxy, SSL termination, virtual hosting via SNI
- **Let's Encrypt** — SSL for all four domains, auto-renews
- **ipwho.is** — free geo/ISP lookup, no API key required, results cached 1hr
- **systemd** — process management, auto-restart on crash

## Running it

```bash
go build -o doihaveipv6 .
PORT=3000 ./doihaveipv6
```

The server reads the client IP from the `X-Real-IP` header (set by nginx). Run it directly and it reads `RemoteAddr` instead, which works fine for local testing.

### API

```
GET /api
```
```json
{ "hasIpv6": true, "ip": "2001:db8::1", "type": "IPv6" }
```

```
GET /geo?ip=<address>
```
```json
{ "isp": "Verizon Business", "city": "Ashburn", "country": "United States" }
```

## DNS setup

For the dual-probe to work, the subdomain DNS records must be strictly single-type:

| Name | Type | Value | Note |
|------|------|-------|------|
| `doihaveipv6.com` | A | your IPv4 | main site |
| `doihaveipv6.com` | AAAA | your IPv6 | main site |
| `ipv4.doihaveipv6.com` | A | your IPv4 | **no AAAA record** |
| `ipv6.doihaveipv6.com` | AAAA | your IPv6 | **no A record** |

No wildcard `*` records — they bleed through and break the probe filtering.

## License

MIT
