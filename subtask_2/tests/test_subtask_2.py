import requests
from time import sleep
import socket

BASE_URL = "http://localhost:8080/access/vs"

# test authentication
def test_authentication():
    # try with no credentials
    response = requests.get(BASE_URL)
    print(response.status_code)
    assert response.status_code == 401

    # try with wrong credentials
    response = requests.get(BASE_URL, auth=("bal", "1fourall"))
    print(response.status_code)
    assert response.status_code == 401

    # try with correct credentials
    response = requests.get(BASE_URL, auth=("bal", "2fourall"))
    print(response.status_code)
    assert response.status_code == 200

# test certificate generation
def test_certificate_generation():
    # create a new vs
    payload = {
        "port": 8443,
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
    sleep(20)

    # test vs resp on http endpoint
    response = requests.get("http://localhost:8443")
    print(response.text)
    assert response.status_code == 200

    # create and add a new certificate on port 8443 vs
    payload = {
		"commonName": "testCert",
        "port": 8443,
        "days": 365
	}

    response = requests.post(BASE_URL + "/certificates/generate", json=payload, auth=("bal", "2fourall"))
    print(response.status_code)
    print(response.json())
    assert response.status_code == 201

    # test vs resp on https endpoint
    response = requests.get("https://localhost:8443", verify=False)
    print(response.text)
    assert response.status_code == 200

# test get all certificates
def test_get_all_certificates():
    response = requests.get(BASE_URL + "/certificates", auth=("bal", "2fourall"))
    print(response.status_code)
    print(response.json())
    assert response.status_code == 200
    assert "8443" in response.json()

# test renew certificate
def test_renew_certificate():
    response = requests.post(BASE_URL + "/certificates/renew/8443", auth=("bal", "2fourall"))
    print(response.status_code)
    print(response.json())
    assert response.status_code == 200
    assert response.json()["port"] == "8443"
    assert response.json()["status"] == "renewed"

# test ip blacklisting
def test_ip_blacklisting():
    container_ip = socket.gethostbyname(socket.gethostname())
    print(container_ip)
    payload = {
        "ip": f"{container_ip}",
        "rule": "block"
    }

    response = requests.post(BASE_URL + "/ip-rules", json=payload, auth=("bal", "2fourall"))
    print(response.status_code)
    print(response.json())
    assert response.status_code == 200
    assert response.json()["ip"] == f"{container_ip}"
    assert response.json()["action"] == "block"
    assert response.json()["status"] == "success"

    sleep(5)

    # test blocked ip
    session = requests.Session()
    adapter = requests.adapters.HTTPAdapter()

    # Bind the request to the container's IP
    adapter.init_poolmanager(
        connections=1,
        maxsize=1,
        source_address=(container_ip, 0)  # Explicitly set source IP
    )
    session.mount("http://", adapter)

    response = session.get(BASE_URL, auth=("bal", "2fourall"))

    print(response.status_code, response.text)
    assert response.status_code == 403
    assert "Forbidden: Your IP has been blocked" in response.text

