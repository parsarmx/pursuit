from pursuit.app import get, post


@get("/hello")
def hello(body):
    # body is dict of json body (or empty)
    name = body.get("name", "world")
    return {"message": f"Hello, {name}!"}


@post("/echo")
def echo(body):
    return {"you_sent": body}
