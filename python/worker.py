import sys
import json
import importlib


def send_response(obj):
    # write JSON on stdout newline-terminated, flush immediately
    sys.stdout.write(json.dumps(obj, separators=(",", ":")) + "\n")
    sys.stdout.flush()


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 worker.py <module_name>", file=sys.stderr)
        sys.exit(1)
    module_name = sys.argv[1]
    # import the user application module (e.g. example_app)
    try:
        user_mod = importlib.import_module(module_name)
    except Exception as e:
        print("Failed to import module", module_name, e, file=sys.stderr)
        sys.exit(1)

    # Resolve ROUTES from pursuit.app (user apps should use pursuit.app decorators)
    import pursuit.app as pa

    # main loop: read newline-delimited JSON on stdin
    while True:
        line = sys.stdin.readline()
        if not line:
            print("kir?")
            break
        line = line.strip()
        if not line:
            continue
        try:
            req = json.loads(line)
            path = req.get("path", "")
            method = req.get("method", "GET")
            body = req.get("body", {})
            headers = req.get("headers", {})

            key = (method.upper(), path)
            handler = pa.ROUTES.get(key)
            if handler is None:
                res = {"status": 404, "body": {"error": "not found"}}
                send_response(res)
                continue

            # call handler, support dict return or (status, dict)
            try:
                out = handler(body)
                if isinstance(out, tuple) and len(out) == 2 and isinstance(out[0], int):
                    status, body_out = out
                else:
                    status, body_out = 200, out
                res = {"status": status, "body": body_out}
            except Exception as e:
                res = {"status": 500, "body": {"error": str(e)}}
            send_response(res)
        except Exception as e:
            send_response(
                {"status": 500, "body": {"error": "bad request", "detail": str(e)}}
            )
