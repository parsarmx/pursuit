import sys, json


def handle(req):
    path = req.get("path")
    method = req.get("method")
    body = req.get("body", {})

    if path == "/hello" and method == "GET":
        return {
            "status": 200,
            "body": {"message": f"Hello {body.get('name', 'world')}"},
        }
    else:
        return {"status": 404, "body": {"error": "Not found"}}


for line in sys.stdin:
    req = json.loads(line)
    resp = handle(req)
    print(json.dumps(resp), flush=True)
