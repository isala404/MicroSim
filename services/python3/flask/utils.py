from models import Route, Response, Fault
import requests
from faults.latency import Latency
from faults.memory_leak import MemoryLeak


def call_next_destination(route: Route) -> Response:
    res = requests.post(route["designation"], json=route, headers={"content-type": "application/json"})
    return Response.from_dict(res.json())


def cast_and_execute(fault: Fault):
    classes = {
        "latency": Latency,
        "memory-leak": MemoryLeak,
    }
    try:
        fault.__class__ = classes[fault.type]
        fault.run()
        return None
    except KeyError:
        return f"fault type {fault.type}, is not implemented"
