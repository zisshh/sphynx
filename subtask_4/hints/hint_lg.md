Content-based routing involves dynamically selecting the backend server based on the request headers and routing rules stored in the ContentRoutingRules attribute of the VirtualService. Follow these steps:

Define Routing Rules:
Store rules like Key: X-Country, Value: US, ServerName: Server3 in the ContentRoutingRules array of the VirtualService.
Implement Server Selection:
Iterate over the rules and match the request headers:

```go
for _, rule := range vs.ContentRoutingRules {
    if req.Header.Get(rule.Key) == rule.Value {
        return findServerByName(vs.ServerList, rule.ServerName)
    }
}
```

Error Handling:
If no match is found or the server is unavailable, log the issue and return HTTP 503 Service Unavailable:
```go
http.Error(res, "No matching rule or healthy servers", http.StatusServiceUnavailable)
```

Dynamic Rule Management:
Use API endpoints (POST, GET, DELETE) to add, retrieve, and delete routing rules.