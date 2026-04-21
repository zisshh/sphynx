FROM golang:1.23

# Install Python, venv, and Redis
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    redis-server

WORKDIR /app

# Copy all test files and solution
COPY . /app/

# Create and activate virtual environment, then install dependencies
RUN python3 -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"
RUN pip3 install -r requirements.txt

# Make scripts executable
RUN chmod +x /app/test_runner.sh
RUN chmod +x /app/server.py
RUN chmod +x /app/start_backend_servers.py
RUN chmod +x /app/test_subtask_4.py

# Initialize Go module and install dependencies
RUN cd /app/solution && go mod tidy

# Start Redis server, log IP and port to a text file
RUN echo "Starting Redis server..." && \
    redis-server --daemonize yes && \
    sleep 2 && \
    REDIS_IP=$(hostname -I | awk '{print $1}') && \
    REDIS_PORT=6379 && \
    echo "Redis server is running on IP: $REDIS_IP, Port: $REDIS_PORT" > /app/redis_server_info.txt

# Define the default command
CMD ["/app/test_runner.sh"]
