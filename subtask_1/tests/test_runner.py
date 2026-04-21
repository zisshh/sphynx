import json
import traceback
import sys
from test_subtask_1 import *

TEST_CASES = [
    {
        "function": test_create_vs,
        "title": "Create Virtual Service",
        "description": "Tests the API endpoint for creating a new virtual service with basic configuration",
        "type": "e2e"
    },
    {
        "function": test_get_all_vs,
        "title": "List Virtual Services",
        "description": "Tests the API endpoint for retrieving all configured virtual services",
        "type": "integration"
    },
    {
        "function": test_get_one_vs,
        "title": "Get Single Virtual Service",
        "description": "Tests the API endpoint for retrieving a specific virtual service by ID",
        "type": "integration"
    },
    {
        "function": test_update_vs,
        "title": "Update Virtual Service",
        "description": "Tests the API endpoint for updating an existing virtual service configuration",
        "type": "e2e"
    },
    {
        "function": test_delete_vs,
        "title": "Delete Virtual Service",
        "description": "Tests the API endpoint for removing a virtual service",
        "type": "e2e"
    },
    {
        "function": test_virtual_service,
        "title": "Virtual Service Functionality",
        "description": "Tests if virtual service properly forwards requests to backend servers",
        "type": "e2e"
    },
    {
        "function": test_round_robin,
        "title": "Round Robin Algorithm",
        "description": "Tests round robin load balancer for equal request distribution among servers",
        "type": "e2e"
    },
    {
        "function": test_weighted_round_robin,
        "title": "Weighted Round Robin Algorithm",
        "description": "Tests weighted round robin load balancer for proportional request distribution",
        "type": "e2e"
    }
]

def execute_test(test_case):
    """
    Executes a single test and returns the result in the required format.
    """
    result = {
        "title": test_case["title"],
        "description": test_case["description"],
        "type": test_case["type"],
        "isPassing": True
    }
    
    try:
        test_case["function"]()
    except AssertionError as e:
        result["isPassing"] = False
        result["error"] = str(e)
        result["notes"] = "Test failed: Assertion error"
    except Exception as e:
        result["isPassing"] = False
        result["error"] = str(e)
        result["notes"] = f"Test failed: Unexpected error\n{traceback.format_exc()}"
    
    return result

def save_results(results):
    """
    Saves test results to test_results.json file.
    """
    try:
        with open('test_results.json', 'w') as f:
            json.dump(results, f, indent=2)
        return True
    except Exception as e:
        print(f"Failed to save results: {e}")
        return False

def main():
    results = []
    all_passed = True
    
    for test_case in TEST_CASES:
        print(f"\nRunning test: {test_case['title']}")
        result = execute_test(test_case)
        results.append(result)
        
        if not result["isPassing"]:
            all_passed = False
            print(f"❌ Test failed: {test_case['title']}")
            if "error" in result:
                print(f"Error: {result['error']}")
            if "notes" in result:
                print(f"Notes: {result['notes']}")
        else:
            print(f"✅ Test passed: {test_case['title']}")
    
    if save_results(results):
        print("\nTest execution completed")
        print("Results written to test_results.json")
    else:
        print("\nFailed to save test results")
        sys.exit(1)
    
    if not all_passed:
        sys.exit(1)
    
    sys.exit(0)

if __name__ == "__main__":
    main()