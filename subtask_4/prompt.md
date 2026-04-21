# Subtask 4: High Throughput Content-Based Routing

---

### Implementation

#### Application/Code Behavior

1. **Content-Based Routing**:
   - Routing functionality directs requests to specific servers based on custom-defined criteria using headers.
   - Administrators can dynamically define routing rules.
   - The scheduling method for the vs should be `content_based`
   - Input Example:
     - JSON configuration specifying routing rules:
       ```json
       {
         "content_routing_rules": [
           { "key": "X-Country", "value": "US", "serverName": "Server3" },
           { "key": "X-Country", "value": "EU", "serverName": "Server4" },
           { "key": "X-Country", "value": "ASIA", "serverName": "Server5" }
         ]
       }
       ```
   - Output Example:
     - A request with the header `X-Country: US` is forwarded to `Server3`.
     - A request with the header `X-Country: EU` is forwarded to `Server4`.
     - A request with the header `X-Country: ASIA` is forwarded to `Server5`.
     - A request with a non-matching header or without the `X-Country` header returns HTTP `503 Service Unavailable` with the message: `"No matching rule or healthy servers"`.

2. **Dynamic Rule Management via API**:
   - Rules can be added, updated, or deleted dynamically through dedicated endpoints.
   - **Endpoints**:
     - **POST** `/access/vs/{vs_id}/rules`: Add a new rule.
       - Input: JSON payload with the rule details.
        ```json
         [
           { "key": "X-Country", "value": "US", "serverName": "Server3" },
           { "key": "X-Country", "value": "EU", "serverName": "Server4" }
         ]
         ```
       - Output:
         ```json
         {
           "status": "success",
           "port": "8443"
         }
         ```
     - **GET** `/access/vs/{vs_id}/rules`: Retrieve all rules for a virtual service.
       - Output:
         ```json
         [
           { "key": "X-Country", "value": "US", "serverName": "Server3" },
           { "key": "X-Country", "value": "EU", "serverName": "Server4" }
         ]
         ```
     - **DELETE** `/access/vs/{vs_id}/rules/{rule_index}`: Delete a specific rule.
       - Output: HTTP `204 No Content`.

3. **Load Balancing Integration**:
   - Integrated load balancing mechanisms like Weighted Round-Robin and Least Connections.
   - Used the `forwardRequest` function to dynamically select the server based on the request context and routing rules.

4. **Default Behavior**:
   - Requests that do not match any rule return HTTP `503 Service Unavailable` with the message: `"No matching rule or healthy servers"`.

#### Code Implementation

- Routing rules are defined in the `content_routing_rules` attribute of a `VirtualService`.
- The `forwardRequest` function determines the appropriate server using the `GetHealthyServer` method, which handles content-based routing logic dynamically.
- The `Start` and `StartHTTPS` methods in `VirtualService` initialize HTTP/HTTPS servers, using `forwardRequest` to process and forward requests.

#### Error Handling

- Logs errors for unmatched rules or misconfigured servers.
- Validates dynamic inputs to ensure correctness.
- Returns meaningful HTTP error codes and messages for invalid requests or configuration mismatches.

#### Design Patterns

1. **Strategy Pattern**:
   - Used for implementing different load balancing algorithms (e.g., Weighted Round-Robin, Least Connections).

2. **Observer Pattern**:
   - Dynamically updates routing rules without requiring a restart.

3. **Chain of Responsibility**:
   - Processes routing rules sequentially to determine the target server.

#### Technical Specifications

1. **Routing Rules Configuration**:
   - Attributes:
     - Key (String): Header name (e.g., `X-Country`).
     - Value (String): Expected value of the header.
     - ServerName (String): The server to route the request to.
   - Input Example:
     ```json
     {
       "content_routing_rules": [
         { "key": "X-Country", "value": "US", "serverName": "Server3" },
         { "key": "X-Country", "value": "EU", "serverName": "Server4" }
       ]
     }
     ```

2. **Endpoints**:
   - **POST** `/access/vs/{vs_id}/rules`: Add a new rule.
     - Input Example:
       ```json
       {
         "key": "X-Country",
         "value": "US",
         "serverName": "Server3"
       }
       ```
     - Output Example:
       ```json
       {
         "status": "success",
         "port": "8443"
       }
       ```
   - **GET** `/access/vs/{vs_id}/rules`: Retrieve all current rules.
     - Output Example:
       ```json
       [
         { "key": "X-Country", "value": "US", "serverName": "Server3" },
         { "key": "X-Country", "value": "EU", "serverName": "Server4" }
       ]
       ```
   - **DELETE** `/access/vs/{vs_id}/rules/{rule_index}`: Delete a specific rule.
     - Output: HTTP `204 No Content`.

3. **Default Behavior**:
   - Requests that do not match any rule:
     - Return HTTP `503 Service Unavailable` with the message: `"No matching rule or healthy servers"`.
     - Log errors for unmatched rules.

NOTE: go module name should be load-balancer for this project
