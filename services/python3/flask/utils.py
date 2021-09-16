from models import Route, Response
import requests


def call_next_destination(route: Route) -> Response:
    res = requests.post(route["designation"], json=route, headers={"content-type": "application/json"})
    return Response.from_dict(res.json())
