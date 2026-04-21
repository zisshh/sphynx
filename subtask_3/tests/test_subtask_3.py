import requests
from time import sleep
from concurrent.futures import ThreadPoolExecutor
from requests.auth import HTTPBasicAuth
from util_request_sender import *

BASE_URL = "http://localhost:8080/access/vs"

# test rate limit feature
def test_rate_limit():
    # create a new vs
    payload = {
        "port": 8010,
        "algorithm": "round_robin",
        "serverList": [
            {
                "name": "Server1",
                "url": "http://localhost:5001",
                "weight": 1
            },
            {
                "name": "Server2",
                "url": "http://localhost:5002",
                "weight": 1
            }
        ]
    }

    response = requests.post(BASE_URL, json=payload, auth=("bal", "2fourall"))
    print(response.status_code)
    assert response.status_code == 201
    sleep(10)

    # set rate limit
    payload = {
        "port": 8010,
        "rate_limit": 10,
        "status_code": 429,
        "message": "Rate limit exceeded."
    }

    response = requests.post(f'{BASE_URL}/rate-limits', json=payload, auth=("bal", "2fourall"))
    print(response.status_code)
    print(response.text)
    assert response.status_code == 200

    # make 15 concurrent requests and check if rate limit is working
    success_count, rate_limit_count = send_request(15)

    print(f"Success: {success_count}")
    print(f"Rate Limit: {rate_limit_count}")

    assert success_count == 10
    assert rate_limit_count == 5

    
    

