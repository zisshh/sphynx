# Load Sphynx — Demo Video Script

Target length: **4 to 6 minutes**.
Layout: put a terminal on the left half, browser on the right half. Keep both visible the whole time so viewers see commands and the UI react together.

---

## STEP 0 — Setup (do this BEFORE you hit record)

Open two terminals and the browser, side-by-side:

- **Terminal A** (the "server" terminal): leave empty for now.
- **Terminal B** (the "client" terminal): where you'll run curl commands.
- **Browser**: open a fresh tab, don't navigate yet.

Check Redis is running (silent prep):
```bash
redis-cli ping
# should print: PONG
```

If it says "not found" or errors out:
```bash
brew install redis && brew services start redis
```

Now start recording.

---

## STEP 1 — Intro (10 seconds)

**DO:** show your face or just the screen with the project folder open in Finder.

**SAY:**
> "Hey, I'm Thakur Divyansh, and this is Load Sphynx — a software load balancer I built in Go. A load balancer is basically a traffic cop for websites: when a bunch of users hit a site, the balancer spreads those requests across multiple backend servers so none of them get overloaded. Let me show you what it does."

---

## STEP 2 — Show the project structure (15 seconds)

**DO:** in Terminal A, run:
```bash
cd "~/Downloads/Code & Projects/project-bloom-main/final_solution"
ls
```

**SAY:**
> "This is the project. It's written in Go. The main logic lives in folders like `loadbalancing`, `healthcheck`, and `handlers`. The `frontend` folder is the dashboard — an admin UI I added on top so you can see everything in one place. There's also a `demo` folder with a helper script that boots five fake backend servers for me."

---

## STEP 3 — Start the whole stack with one command (20 seconds)

**DO:** in Terminal A, run:
```bash
./demo/run_demo.sh
```

Wait about 3 seconds until you see the line `Dashboard UI: http://localhost:8080/`.

**SAY:**
> "I wrote a single script that does three things: it checks Redis is up, it boots five small backend servers on ports 5001 through 5005, and then it starts Load Sphynx itself. In real life these backends would be your actual app servers — here they're just tiny placeholders that reply with their name so we can see where our traffic ended up."

---

## STEP 4 — Open the dashboard (20 seconds)

**DO:** in the browser, go to:
```
http://localhost:8080/
```

Wait a second for it to load.

**SAY:**
> "This is the admin dashboard. At the top I can see it's already picked up two virtual services — think of a virtual service as one public endpoint that hides a group of backend servers. Five servers are healthy, and the health check column is green. The balancer is actively pinging each backend every couple of seconds, and if one goes down it'll notice."

---

## STEP 5 — Virtual Services tab (30 seconds)

**DO:** click **Virtual Services** in the left sidebar. Then click the blue **eye icon** on the row for port 8001.

**SAY:**
> "Under Virtual Services I can see the two endpoints I've configured. Port 8001 is running round-robin across two backends, and port 8443 is running weighted round-robin across three backends."
>
> "Let me open the first one. Round-robin just means every new request goes to the next server in the list — so request one goes to Server1, request two goes to Server2, request three goes back to Server1, and so on. It's the simplest way to spread load evenly."

Close the modal with the **X** or press Escape.

---

## STEP 6 — Prove round-robin works from the terminal (30 seconds)

**DO:** in Terminal B, run:
```bash
for i in 1 2 3 4 5 6; do curl -s http://localhost:8001/; done
```

**SAY:**
> "Let me prove that. I'm going to send six requests in a row to port 8001. Watch the responses."

Let the output print. It should alternate `Server1 / Server2 / Server1 / Server2 / Server1 / Server2`.

> "Perfect. Server1, Server2, Server1, Server2 — the balancer is flipping between the two backends on every single request. That's round-robin in action."

---

## STEP 7 — Weighted round-robin (30 seconds)

**DO:** in Terminal B:
```bash
for i in {1..10}; do curl -s http://localhost:8443/; done
```

**SAY:**
> "Port 8443 is running a weighted algorithm. In my config I gave Server3 a weight of 2, Server4 a weight of 1, and Server5 a weight of 2. That means Server3 and Server5 should each get roughly twice as much traffic as Server4 over time. This is useful when some of your machines are beefier than others — you send them more work."

Point at the output. Roughly two of every five lines should be Server4.

> "And you can see — Server4 only shows up about once for every two Server3s and Server5s. Weights are respected."

---

## STEP 8 — Kill a backend, watch the balancer react (45 seconds)

**DO:** in Terminal B:
```bash
lsof -ti:5001 | xargs kill
```

Then go back to the browser, click **Dashboard** in the sidebar, and click **Refresh** (top right). Wait ~2 seconds.

**SAY:**
> "Okay, now the fun part — let me simulate a server dying. I'm killing the process on port 5001, which is Server1."
>
> "Now I'll flip back to the dashboard and refresh. Look at that — the port 8001 row changed. It used to say 'two out of two healthy'. Now it's one out of two, and the status badge flipped from green to yellow. The balancer noticed within a couple of seconds because it's pinging each backend's health endpoint on a loop."

**DO:** in Terminal B:
```bash
for i in 1 2 3 4 5 6; do curl -s http://localhost:8001/; done
```

**SAY:**
> "And if I send traffic now — every single request goes to Server2. Server1 has been automatically taken out of the rotation. No code change, no restart. The balancer is self-healing."

---

## STEP 9 — Bring Server1 back (15 seconds)

**DO:** in Terminal A (or a new terminal tab), run:
```bash
cd "~/Downloads/Code & Projects/project-bloom-main/final_solution"
go run demo/backends.go 5001 Server1 &
```

Refresh the dashboard.

**SAY:**
> "And if I bring Server1 back online — I'll just relaunch the process — and refresh the dashboard, look at that. Two out of two healthy again. It rejoined the pool automatically."

---

## STEP 10 — IP Blacklisting (30 seconds)

**DO:** click **IP Blacklisting** in the sidebar. Click the red **Block IP** button. Type `203.0.113.42` and save.

**SAY:**
> "Load Sphynx also has security features. If I'm getting abusive traffic from a specific IP, I can block it right from the UI. Let me blacklist this address — it's a fake one just for the demo. Now that IP is on a deny list, and every request from it will get rejected at the front door, before it even hits my backends."

---

## STEP 11 — Rate Limiting (30 seconds)

**DO:** click **Rate Limiting** in the sidebar. Click **Edit** on the row for port 8001. Change the limit from 15 to 50. Save.

**SAY:**
> "Next up, rate limiting. This protects me from anyone trying to hammer my servers with too many requests per minute. Right now port 8001 is capped at fifteen requests per minute — after fifteen, everyone else gets a 429 Too Many Requests error. I can change that live from the dashboard without restarting anything. Let me bump it to fifty."
>
> "Behind the scenes, Load Sphynx is storing those counters in Redis, which is why it works even if I had multiple copies of the balancer running."

---

## STEP 12 — SSL Certificates (20 seconds)

**DO:** click **SSL Certificates** in the sidebar.

**SAY:**
> "The dashboard also lets me manage TLS certificates per virtual service. Right now I'm running everything over plain HTTP to keep the demo simple, but in production I could generate a self-signed cert or upload a real one from Let's Encrypt, attach it to a virtual service, and that endpoint instantly becomes HTTPS."

---

## STEP 13 — Content Routing (30 seconds)

**DO:** click **Content Routing** in the sidebar. Click the Virtual Service dropdown, pick port 8001.

**SAY:**
> "Last feature — content-based routing. This lets me send specific kinds of requests to specific backends. For example, I could say: any request with the header `X-User-Tier: premium` goes to Server1, and everyone else goes to Server2. That's how you'd do things like A/B tests, canary releases, or tenant isolation. You set the rule here in the UI and Load Sphynx does the matching in real time."

---

## STEP 14 — Wrap up (20 seconds)

**DO:** go back to the Dashboard tab. Let it sit there.

**SAY:**
> "So that's Load Sphynx. Built in Go, runs on a single binary, gives me round-robin, weighted, and least-connections algorithms, automatic health checks, rate limiting backed by Redis, IP blacklisting, TLS, and content routing — all from a clean dashboard. In my benchmarks it handled around twelve thousand requests per second at a 35 millisecond p99 latency, which puts it in the same ballpark as tools like Nginx and HAProxy, but with a lot less configuration overhead."
>
> "Thanks for watching."

---

## Cheat sheet — all the commands in one place

Keep this open in a note while recording so you don't have to remember:

```bash
# Start everything
cd "~/Downloads/Code & Projects/project-bloom-main/final_solution"
./demo/run_demo.sh

# Prove round-robin (VS 8001)
for i in 1 2 3 4 5 6; do curl -s http://localhost:8001/; done

# Prove weighted round-robin (VS 8443)
for i in {1..10}; do curl -s http://localhost:8443/; done

# Kill Server1
lsof -ti:5001 | xargs kill

# Bring Server1 back
go run demo/backends.go 5001 Server1 &

# Teardown
lsof -ti:8080,8001,8443,5001,5002,5003,5004,5005 | xargs kill -9
```

Dashboard URL: **http://localhost:8080/**

---

## Tips for smooth recording

1. **Rehearse once.** Run through the whole script with no narration, just the commands. Fix any port conflicts before you hit record.
2. **Increase terminal font size.** 16–18 pt. Small fonts look unprofessional on video.
3. **Clear the terminal before each section** with `clear` — less distracting.
4. **Pause the balancer's health-check logs** if they're too noisy. Terminal A will keep printing; just let it roll or minimize it between sections.
5. **Record at 1080p or higher** (`⌘ + Shift + 5` → Options → Screen Quality).
6. **Don't read word-for-word.** Skim a section, then talk naturally. It'll feel alive instead of robotic.
7. **If something breaks mid-record**, keep going — you can edit it out. Don't restart the whole take.

Good luck.
