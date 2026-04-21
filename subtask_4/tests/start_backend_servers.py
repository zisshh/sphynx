import threading
import server

servers = [
        ("Server1", 5001),
        ("Server2", 5002),
        ("Server3", 5003),
        ("Server4", 5004),
        ("Server5", 5005)
    ]

threads = []
for name, port in servers:
    thread = threading.Thread(target=server.create_app, args=(name, port))
    threads.append(thread)
    thread.start()