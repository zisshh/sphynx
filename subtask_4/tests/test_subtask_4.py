import requests
from time import sleep

BASE_URL = "http://localhost:8080/access/vs"

# test adding routing rules
def test_add_routing_rules():
    # create a new vs
    payload = {
        "port": 8090,
        "algorithm": "content_based",
        "rate_limit": 50,
        "status_code": 429,
        "message": "Too many requests - Try again later.",
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
            },
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

    response = requests.post(BASE_URL, json=payload, auth=("bal", "2fourall"))
    print(response.status_code)
    assert response.status_code == 201
    sleep(10)

    # add routing rules
    payload = {
        "key": "X-Country",
        "value": "US",
        "serverName": "Server3"
    }

    response = requests.post(BASE_URL + "/8090/rules", json=payload, auth=("bal", "2fourall"))
    print(response.status_code)
    print(response.json())
    assert response.status_code == 201

    # test routing 
    response = requests.get("http://localhost:8090", headers={"X-Country": "US"})
    print(response.text)
    assert response.status_code == 200
    assert response.text == "Server3"

# test get routing rules
def test_get_routing_rules():
    response = requests.get(BASE_URL + "/8090/rules", auth=("bal", "2fourall"))
    print(response.status_code)
    print(response.json())
    assert response.status_code == 200
    assert response.json() == [{"key": "X-Country", "value": "US", "serverName": "Server3"}]

# test delete routing rules
def test_delete_routing_rules():
    response = requests.delete(BASE_URL + "/8090/rules/0", auth=("bal", "2fourall"))
    print(response.status_code)
    assert response.status_code == 204

    response = requests.get(BASE_URL + "/8090/rules", auth=("bal", "2fourall"))
    print(response.status_code)
    print(response.json())
    assert response.status_code == 200
    assert response.json() == [] # empty list

    