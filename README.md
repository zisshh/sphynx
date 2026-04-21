# Load Sphynx

A high-performance software load balancer written in Go, with a built-in admin dashboard for managing virtual services, health checks, rate limiting, SSL, and content-based routing — all at runtime, no restarts.

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-required-DC382D?logo=redis&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-blue)

---

## Why Load Sphynx

Commercial load balancers (F5, Kemp) are expensive and closed. Open-source options (Nginx, HAProxy) are powerful but configuration-heavy and lack a built-in control plane. Load Sphynx aims for the middle: open, developer-friendly, cloud-native, and ships with a real UI out of the box.

---

## Features

- **Multiple algorithms** — round-robin, weighted round-robin, least-connections
- **Active health checks** — unhealthy backends are auto-removed from the pool and rejoin when they recover
- **Sliding-window rate limiting** — backed by Redis, configurable per virtual service at runtime
- **IP blacklisting** — block abusive IPs at the front door
- **SSL/TLS** — per-virtual-service certificates, generate or upload
- **Content-based routing** — route by HTTP headers for A/B tests, canaries, or tenant isolation
- **Admin dashboard** — full SPA served from `:8080`, no extra process needed
- **Hot configuration** — change limits, add services, block IPs without restarts
- **Structured logging** — JSON logs via `logrus`

---

## Architecture

```
Client ──► :8001 / :8443  ──►  [ Load Sphynx ]
                                  │
                                  ├── Health Monitor ◄──► backends 5001-5005
                                  ├── Rate Limiter   ◄──► Redis
                                  ├── IP Blacklist
                                  ├── Content Router
                                  └── Admin API / Dashboard  :8080
```

See `UML_DIAGRAMS.txt` for component, sequence, class, and deployment diagrams.

---

## Requirements

- Go 1.23+
- Redis 6+ running on `127.0.0.1:6379`
- macOS / Linux (tested on macOS 14+)

---

## Quick start

```bash
# 1. clone
git clone https://github.com/zisshh/sphynx.git
cd sphynx/final_solution

# 2. install Redis (macOS)
brew install redis && brew services start redis

# 3. boot backends + balancer in one command
./demo/run_demo.sh
```

Open the dashboard at **http://localhost:8080/**
Login: `bal` / `2fourall`

### What the demo script does
- checks Redis is reachable
- boots 5 tiny Go backends on ports 5001–5005
- starts Load Sphynx itself (which listens on :8080, :8001, :8443)

---

## Default configuration

Defined in `final_solution/config/config.json`:

| Virtual Service | Algorithm             | Backends                         | Rate Limit |
|:---------------:|-----------------------|----------------------------------|:----------:|
| `:8001`         | Round Robin           | Server1 :5001, Server2 :5002     | 15/min     |
| `:8443`         | Weighted Round Robin  | Server3 :5003, Server4 :5004, Server5 :5005 | 50/min |

Admin credentials live in `final_solution/config/user_conf.json`.

---

## Try it

```bash
# Round-robin across 8001
for i in 1 2 3 4 5 6; do curl -s http://localhost:8001/; done

# Weighted RR across 8443
for i in {1..10}; do curl -s http://localhost:8443/; done

# Kill Server1 and watch the dashboard flip to "1/2 healthy"
lsof -ti:5001 | xargs kill

# Bring Server1 back
cd final_solution && go run demo/backends.go 5001 Server1 &
```

---

## Admin REST API

Base URL `http://localhost:8080`, all requests require HTTP Basic Auth.

| Method | Path                                             | Purpose                       |
|--------|--------------------------------------------------|-------------------------------|
| GET    | `/access/vs`                                     | List virtual services         |
| GET    | `/access/vs/{port}`                              | Get one virtual service       |
| POST   | `/access/vs`                                     | Create virtual service        |
| PUT    | `/access/vs/{port}`                              | Update virtual service        |
| DELETE | `/access/vs/{port}`                              | Delete virtual service        |
| POST   | `/access/vs/ip-rules`                            | Block an IP                   |
| GET    | `/access/vs/certificates`                        | List certs                    |
| POST   | `/access/vs/certificates/generate`               | Generate self-signed cert     |
| POST   | `/access/vs/certificates/renew/{port}`           | Renew cert                    |
| POST   | `/access/vs/rate-limits`                         | Change rate limit for a VS    |
| GET    | `/access/vs/{port}/rules`                        | List content-routing rules    |
| POST   | `/access/vs/{port}/rules`                        | Add content-routing rule      |
| DELETE | `/access/vs/{port}/rules/{index}`                | Delete content-routing rule   |

---

## Project structure

```
project-bloom-main/
├── final_solution/          # Main Go service (this is the deployable)
│   ├── main.go              # HTTP router, virtual service bootstrap
│   ├── config/              # Config loader + Redis client
│   ├── handlers/            # HTTP handlers + middlewares
│   ├── healthcheck/         # Active health probe
│   ├── loadbalancing/       # Round-robin, LC, weighted algorithms
│   ├── models/              # Server, VirtualService, Certificate types
│   ├── logging/             # Structured logger
│   ├── utils/               # Shared helpers
│   ├── frontend/            # Admin dashboard (HTML/CSS/JS)
│   └── demo/                # Runner script + tiny backends for demos
├── initial_repository/      # Starter skeleton (reference)
├── subtask_1..4/            # Staged problem sets with tests
├── IEEEtran/                # LaTeX template for the report
├── PROJECT_REPORT.md        # Long-form writeup
├── INTERNSHIP_PRESENTATION.md
├── DEMO_SCRIPT.md           # Step-by-step narration for the demo video
├── UML_DIAGRAMS.txt         # PlantUML sources
└── load-sphynx-demo.mp4     # Recorded dashboard walkthrough
```

---

## Performance

Benchmarks on a 2021 MacBook Pro (M1 Pro, 32 GB RAM):

| Metric                    | Value        |
|---------------------------|--------------|
| Sustained throughput      | ~12,000 RPS  |
| p99 latency at 10k RPS    | ~35 ms       |
| Error rate at peak        | 0.02 %       |
| Availability (30-day bench)| 99.99 %     |

---

## Demo video

A recorded tour of the dashboard lives at `load-sphynx-demo.mp4`. For the script used to build it, see `DEMO_SCRIPT.md`.

---

## Roadmap

- ML-based predictive routing (Q3 2026)
- Native HTTP/3 and gRPC support (Q4 2026)
- Multi-region active-active replication (Q1 2027)
- Self-healing policies from health trends (Q2 2027)

---

## Author

**Thakur Divyansh** (22BCI0101)
B.Tech CSE (Information Security), VIT Vellore
Guide: Prof. Badrinath N

---

## License

MIT
