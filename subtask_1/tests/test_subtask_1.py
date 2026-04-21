import requests
from time import sleep
import server
import threading

BASE_URL = "http://localhost:8080/access/vs"

# test create endpoint
def test_create_vs():
    payload = {
        "port": 8001,
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

    response = requests.post(BASE_URL, json=payload)
    print(response.status_code)
    assert response.status_code == 201
    sleep(20)

# test get all vs endpoint
def test_get_all_vs():
    response = requests.get(BASE_URL)
    print(response.json())
    assert response.status_code == 200

# test get one vs endpoint
def test_get_one_vs():
    response = requests.get(BASE_URL + "/8001")
    print(response.json())
    assert response.status_code == 200

# test update vs endpoint
def test_update_vs():
    payload = {
        "port": 8001,
        "algorithm": "weighted_round_robin",
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

    response = requests.put(BASE_URL + "/8001", json=payload)
    print(response.status_code)
    assert response.status_code == 200
    sleep(20)

    # check if actually updated
    response_after_update = requests.get(BASE_URL + "/8001")
    print(response_after_update.json())
    assert response_after_update.json()["algorithm"] == "weighted_round_robin"

# test delete vs endpoint
def test_delete_vs():
    response = requests.delete(BASE_URL + "/8001")
    print(response.status_code)
    assert response.status_code == 204

    # check if actually deleted
    response_after_delete = requests.get(BASE_URL + "/8001")
    print(response_after_delete.status_code)
    assert response_after_delete.status_code == 404
    

# check if virtual service is working
def test_virtual_service():
    # create a sample vs
    payload = {
        "port": 8002,
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

    response = requests.post(BASE_URL, json=payload)
    print(response.status_code)
    assert response.status_code == 201
    sleep(5)

    # check if vs is working
    response = requests.get("http://localhost:8002")
    print(response.text)
    assert response.status_code == 200

# test round robin
def test_round_robin():
    # create a sample vs
    payload = {
        "port": 8003,
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

    response = requests.post(BASE_URL, json=payload)
    print(response.status_code)
    assert response.status_code == 201
    sleep(5)

    # check if round robin is working
    responses = []
    for _ in range(10):
        response = requests.get("http://localhost:8003")
        responses.append(response.text)
        # sleep(2)

    print("Responses:", responses)

    # Check if responses are in round robin sequence
    for i in range(len(responses) - 1):
        if responses[i] == "Server1":
            assert responses[i + 1] == "Server2"
        elif responses[i] == "Server2":
            assert responses[i + 1] == "Server1"

# test weighted round robin
def test_weighted_round_robin():
    payload = {
        "port": 8004,
        "algorithm": "weighted_round_robin",
        "serverList": [
            {
                "name": "Server3",
                "url": "http://localhost:5003",
                "weight": 2
            },
            {
                "name": "Server4",
                "url": "http://localhost:5004",
                "weight": 1
            },
            {
                "name": "Server5",
                "url": "http://localhost:5005",
                "weight": 2
            }
        ]
    }

    response = requests.post(BASE_URL, json=payload)
    print(response.status_code)
    assert response.status_code == 201
    sleep(5)

    # check if weighted round robin is working
    responses = []
    for _ in range(10):
        response = requests.get("http://localhost:8004")
        responses.append(response.text)
        # sleep(2)

    server3_count = responses.count("Server3")
    server4_count = responses.count("Server4")
    server5_count = responses.count("Server5")

    print("Server3 count:", server3_count)
    print("Server4 count:", server4_count)
    print("Server5 count:", server5_count)

    assert server3_count == 4
    assert server4_count == 2
    assert server5_count == 4
    
    
# test_create_vs()
# test_get_all_vs()
# test_get_one_vs()
# test_update_vs()
# test_delete_vs()
# test_virtual_service()
# test_round_robin()
# test_weighted_round_robin()

