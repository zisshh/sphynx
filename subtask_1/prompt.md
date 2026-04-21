# Subtask 1: Core Load Balancer Implementation

---

## Requirements

### Application/Code Behavior

1. **Server Health Checks**:
   - Perform periodic health checks on all servers in the server list at 2-second intervals.
   - Servers failing the health check (e.g., returning non-200 HTTP status codes or being unreachable) must be marked as unhealthy.
   - Attempt recovery by rechecking their health status in subsequent cycles.
   - Health check specifics:
     - Send lightweight HTTP HEAD requests to server endpoints.
     - Consider non-200 responses as failures.
     - Retry health checks up to 3 times before marking a server as unhealthy.
     - Log server failures and recovery events.

2. **Load Balancing Algorithms**:
   - Implement these server selection algorithms:
     1. **Round-Robin**: Cycle through healthy servers sequentially.
     2. **Weighted Round-Robin**: Distribute requests based on server weights.
     3. **Least Connections**: Assign requests to the server with the fewest active connections.
     4. **Weighted Least Connections**: Incorporate weights into connection-based balancing.
   - Input:
     - A list of servers, each with attributes: name, URL, weight, health status, and active connections.
   - Output:
     - Details of the selected server (e.g., `{"name": "Server1", "URL": "http://example.com"}`).

3. **VirtualService and Server Classes**:
   - **VirtualService**:
     - Attributes: Port, Algorithm, and a list of Server instances.
     - Behaviors: Manage traffic distribution and perform server health checks.
   - **Server**:
     - Attributes: Name, URL, Weight, Health status, and Active connections.
     - Behaviors: Conduct health checks, update health status, and track connection counts.

4. **Dynamic Virtual Service Initialization**:
   - Support initialization from a JSON file (`config.json`).
   - JSON file format example:
     ```json
     [
       {
         "port": 7001,
         "algorithm": "round_robin",
         "serverList": [
           {"name": "Server1", "url": "http://localhost:5001", "weight": 1},
           {"name": "Server2", "url": "http://localhost:5002", "weight": 1}
         ]
       }
     ]
     ```

5. **REST API Endpoints**:
   - Provide endpoints for managing virtual services:
     - **GET** `/access/vs`: Retrieve all virtual services.
       - Response Status Code: 200 (OK)
       - Response Format: JSON array of virtual services.
       - Input Example: None
       - Output Example:
         ```json
          [
            {
              "port": 7001,
              "algorithm": "round_robin",
              "serverList": [
                {
                  "name": "Server1",
                  "url": "http://localhost:5001",
                  "weight": 1,
                  "health": false,
                  "connections": 0
                },
                {
                  "name": "Server2",
                  "url": "http://localhost:5002",
                  "weight": 1,
                  "health": false,
                  "connections": 0
                }
              ]
            },
            {
              "port": 7002,
              "algorithm": "weighted_round_robin",
              "serverList": [
                {
                  "name": "Server3",
                  "url": "http://localhost:5003",
                  "weight": 2,
                  "health": false,
                  "connections": 0
                },
                {
                  "name": "Server4",
                  "url": "http://localhost:5004",
                  "weight": 1,
                  "health": false,
                  "connections": 0
                },
                {
                  "name": "Server5",
                  "url": "http://localhost:5005",
                  "weight": 2,
                  "health": false,
                  "connections": 0
                }
              ]
            }
          ]
         ```

     - **GET** `/access/vs/{vs_id}`: Retrieve a specific virtual service by port.
       - Input Example:
         ```http
         GET /access/vs/7001 HTTP/1.1
         Host: localhost:8080
         ```
       - Response Status Codes:
         - 200 (OK) on success
         - 404 (Not Found) if the service does not exist.
       - Output Example:
         ```json
          {
            "port": 7001,
            "algorithm": "round_robin",
            "serverList": [
              {
                "name": "Server1",
                "url": "http://localhost:5001",
                "weight": 1,
                "health": false,
                "connections": 0
              },
              {
                "name": "Server2",
                "url": "http://localhost:5002",
                "weight": 1,
                "health": false,
                "connections": 0
              }
            ]
          }
         ```

     - **POST** `/access/vs`: Create a new virtual service.
       - Input Example:
         ```json
         {
           "port": 7002,
           "algorithm": "weighted_round_robin",
           "serverList": [
             {"name": "Server3", "url": "http://localhost:5003", "weight": 2},
             {"name": "Server4", "url": "http://localhost:5004", "weight": 1}
           ]
         }
         ```
       - Response Status Codes:
         - 201 (Created) on success.
         - 400 (Bad Request) for invalid input.

     - **PUT** `/access/vs/{vs_id}`: Update an existing virtual service.
       - Input Example:
         ```json
         {
           "port": 7001,
           "algorithm": "least_connections",
           "serverList": [
             {"name": "Server1", "url": "http://localhost:5001", "weight": 1}
           ]
         }
         ```
       - Response Status Codes:
         - 200 (OK) on successful update.
         - 400 (Bad Request) for invalid input.
         - 404 (Not Found) if the service does not exist.

     - **DELETE** `/access/vs/{vs_id}`: Remove a virtual service.
       - Input Example:
         ```http
         DELETE /access/vs/7001 HTTP/1.1
         Host: localhost:8080
         ```
       - Response Status Codes:
         - 204 (No Content) on successful deletion.
         - 404 (Not Found) if the service does not exist.

6. **Dynamic Configuration Updates**:
   - Support real-time updates to the configuration via API calls.
   - Ensure graceful service shutdown and restart during updates or deletions.

---

### Expected Architectural Decisions

- Encapsulate logic within `VirtualService` and `Server` classes.
- Use modular and reusable components for health checks and load balancing algorithms.
- Manage virtual services thread-safely with REST API handlers.
- Implement independent logging for each virtual service.

---

### Software Design Patterns

- **Singleton Pattern**: For managing shared configurations.
- **Factory Pattern**: For creating virtual services and servers.
- **Strategy Pattern**: For dynamically selecting load balancing algorithms.

---

### Advanced Language Features

- Use concurrency to:
  - Perform periodic health checks.
  - Manage shared virtual service states.
- Implement synchronization mechanisms to ensure thread safety.

---

### Error Handling

- Log errors during health checks or configuration operations.
- Validate all user inputs and API payloads.
- Provide meaningful API responses with appropriate status codes.
- Ensure graceful service shutdown and error recovery during updates or deletions.

---

### Technical Specifications

- **Class Specifications**:
  - `VirtualService`:
    - Attributes: Port, Algorithm, Server list.
    - Methods: Start server, manage traffic, and perform health checks.
  - `Server`:
    - Attributes: Name, URL, Weight, Health status, Connections.
    - Methods: CheckHealth, ForwardRequest.

- **Load Balancing Algorithms**:
  - Implement Round-Robin, Weighted Round-Robin, Least Connections, and Weighted Least Connections.

- **API Details**:
  - Base URL: `/access/vs`
  - Methods: GET, POST, PUT, DELETE.
  - Input/output examples must match the given specifications.

---

NOTE: go module name should be load-balancer for this project
