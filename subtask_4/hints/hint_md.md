To implement content-based routing, focus on inspecting specific HTTP headers in the incoming requests and matching them against a set of dynamically configurable rules in the VirtualService. Use the ContentRoutingRules attribute to store the routing logic. Here’s a small pointer:

```go
if req.Header.Get("X-Country") == "US" {
    forwardToServer("Server3")
}
```

Also, ensure unmatched requests return HTTP 503 Service Unavailable.