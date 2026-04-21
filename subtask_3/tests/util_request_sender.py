import requests
from concurrent.futures import ThreadPoolExecutor
from requests.auth import HTTPBasicAuth

URL = "http://localhost:8080/access/vs/8010"
AUTH = HTTPBasicAuth("bal", "2fourall")  # Set Basic Auth credentials

def send_request(n):
    # Counters for responses
    success_count = 0
    rate_limit_count = 0
    failure_count = 0

    def make_request():
        nonlocal success_count, rate_limit_count, failure_count
        response = requests.get(URL, auth=AUTH)
        if response.status_code == 200:
            success_count += 1
        elif response.status_code == 429:
            print(response.text)
            rate_limit_count += 1
        

    # Simulate n concurrent requests
    with ThreadPoolExecutor(max_workers=10) as executor:
        futures = [executor.submit(make_request) for _ in range(n)]

    # Wait for all futures to complete
    for future in futures:
        future.result()

    return success_count, rate_limit_count


