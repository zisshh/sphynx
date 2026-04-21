import json
import traceback
import sys
from test_subtask_4 import *

TEST_CASES = [
    {
        "function": test_add_routing_rules,  
        "title": "Add routing rules",
        "description": "Adds a routing rule to the load balancer, and tests if the rule is applied correctly",
        "type": "e2e"
    },
    {
        "function": test_get_routing_rules,
        "title": "Get routing rules",
        "description": "Tests if the routing rules are retrieved correctly",
        "type": "integration"
    },
    {
        "function": test_delete_routing_rules,
        "title": "Delete routing rules",
        "description": "Deletes a routing rule from the load balancer, and tests if the rule is removed correctly",
        "type": "integration"
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