ROUTES = {}  # key: (method, path) -> handler function


def route(method: str, path: str):
    def decorator(fn):
        key = (method.upper(), path)
        ROUTES[key] = fn
        return fn

    return decorator


def get(path: str):
    return route("GET", path)


def post(path: str):
    return route("POST", path)
