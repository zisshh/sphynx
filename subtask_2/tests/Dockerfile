FROM golang:1.23

# Install Python and venv
RUN apt-get update && apt-get install -y python3 python3-pip python3-venv

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
RUN chmod +x /app/test_subtask_2.py

# Initialize Go module and install dependencies
RUN cd /app/solution && go mod tidy

CMD ["/app/test_runner.sh"]