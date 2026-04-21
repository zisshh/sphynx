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
RUN chmod +x /app/test_subtask_3.py

# Initialize Go module and install dependencies
RUN cd /app/solution && go mod tidy

# Ensure Redis starts in the foreground and the container waits for it
CMD redis-server --bind 0.0.0.0 --port 6379 & \
    sleep 5 && \
    export REDIS_HOST=127.0.0.1:6379 && \
    echo "Redis started on $REDIS_HOST" && \
    /app/test_runner.sh
