# Subtask 2: SSL and Secure Connections Endpoints:

1. **Basic Authentication**:
   - Middleware is applied to all endpoints, ensuring secure access.
   - Use credentials `bal:2fourall` for implementation and validation.
   - **Input Example**:
     ```http
     GET /access/vs HTTP/1.1
     Authorization: Basic dGhpcyBpcyBhIGJhc2U2NCBzdHJpbmc=
     ```
   - **Response Examples**:
     - For invalid credentials:
       ```json
       {
         "error": "Unauthorized: Invalid credentials"
       }
       ```
       - **HTTP Status Code**: 401 Unauthorized


2. **Certificate Creation**:

   - Provide an endpoint to generate self-signed certificates for a specified virtual service.
   - Automatically attach generated certificates to the respective service.
   - **Endpoint**:
     - **POST** `/access/vs/certificates/generate`: Generate a self-signed certificate.
       - **Input Example**:
         ```json
         {
           "commonName": "example.com",
           "port": 8443,
           "days": 365
         }
         ```
       - **Output Example**:
         ```json
         {
           "status": "success",
           "cert": "/certs/8443/cert.crt",
           "key": "/certs/8443/key.key",
           "port": 8443
         }
         ```
         - **HTTP Status Code**: 201 Created

3. **Certificate Renewal**:

   - Allow renewal of existing certificates through an API endpoint.
   - Certificates nearing expiration are identified and automatically rotated.
   - **Endpoints**:
     - **POST** `/access/vs/certificates/renew/{port}`: Renew an SSL certificate for a specified port.
       - **Output Example**:
         ```json
         {
           "status": "renewed",
           "port": 8443
         }
         ```
         - **HTTP Status Code**: 200 OK

   - Certificates are validated for integrity and backed up before renewal.

4. **Certificate Rotation**:

   - Implement a daily background job to check and renew expiring certificates automatically.
   - Certificates expiring within the next 30 days are prioritized for renewal.

5. **Configure HTTPS Endpoints**:

   - Virtual services with certificates are made available via HTTPS.
   - Automatically redirect HTTP traffic to HTTPS if a certificate is associated with the service.
   - Support backward compatibility for services without SSL certificates.
   - **HTTP Status Codes**:
     - **301 Moved Permanently**: For HTTP-to-HTTPS redirection.
     - **200 OK**: For successful HTTPS responses.

6. **IP Blacklisting**:

   - Enable blocking of specific IP addresses through API.
   - Blocked IPs are denied access to all virtual services.
   - **Endpoints**:
     - **POST** `/access/vs/ip-rules`: Add a rule to block an IP.
       - **Input Example**:
         ```json
         {
           "rule": "block",
           "ip": "192.168.1.1"
         }
         ```
       - **Output Example**:
         ```json
         {
           "status": "success",
           "ip": "192.168.1.1",
           "action": "block"
         }
         ```
         - **HTTP Status Code**: 200 OK

   - Middleware dynamically enforces blacklisting without requiring a service restart.
       - 403 with message "Forbidden: Your IP has been blocked"

---

### Expected Architectural Decisions

- **Secure Storage**:
  - SSL certificates and private keys are stored in a secure, isolated directory.
  - Certificates are validated for integrity and compatibility upon upload.

- **Dynamic Updates**:
  - Changes to IP rules and SSL configurations should take effect without restarting the application.

- **Middleware Architecture**:
  - Implement reusable middleware for IP blacklisting and Basic Authentication.

- **Backward Compatibility**:
  - Services without certificates operate seamlessly over HTTP.

### Software Design Patterns

- **Decorator Pattern**:
  - For middleware managing IP blacklisting and Basic Authentication.
- **Observer Pattern**:
  - For monitoring and dynamically applying updates to IP rules or SSL configurations.
- **Builder Pattern**:
  - For constructing HTTPS-enabled virtual services with associated certificates.

### Advanced Language Features

- Concurrency for handling multiple secure and non-secure connections.
- Secure file access mechanisms for certificates and keys.
- Efficient IP blacklisting using trie-based or regex matching.

### Error Handling

- Provide descriptive error messages for:
  - Missing or invalid SSL certificates.
  - Unauthorized access attempts (e.g., invalid credentials or blacklisted IPs).
  - Malformed API requests.
- Ensure graceful fallback to HTTP if HTTPS setup fails.

### Technical Specifications

1. **Basic Authentication**:
   - Middleware to validate `Authorization` headers for secured endpoints using credentials `bal:2fourall`.

2. **SSL Certificate Management**:
   - **POST** `/access/vs/certificates/upload`: Upload an SSL certificate and key.
   - **POST** `/access/vs/certificates/generate`: Generate a self-signed certificate.
   - **POST** `/access/vs/certificates/renew/{port}`: Renew an existing certificate.

3. **IP Blacklisting**:
   - **POST** `/access/vs/ip-rules`: Add a new rule to block an IP.

4. **HTTPS Configuration**:
   - Automatically redirect HTTP traffic to HTTPS if certificates are available.
   - Dynamically restart virtual services to apply SSL changes.

NOTE: go module name should be load-balancer for this project
