from models import Route, Response, Fault
import requests
from faults.latency import Latency
from faults.memory_leak import MemoryLeak
import logging

logger = logging.getLogger('service')

def call_next_destination(route: Route, req_id: str) -> Response:
    logger.info(f'RequestID={req_id}, Calling Next Destination, Designation={route["designation"]} Body={route}')
    res = requests.post(route["designation"], json=route, headers={"content-type": "application/json", "X-Request-ID": req_id})
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
