from dataclasses import dataclass
from typing import List, Any, Optional
from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class Fault:
    type: str
    args: Any

    def run(self):
        pass


@dataclass_json
@dataclass
class FaultState:
    before: List[Fault]
    after: List[Fault]


@dataclass_json
@dataclass
class Route:
    designation: str
    faults: FaultState
    routes: Any = None


@dataclass_json
@dataclass
class Response:
    service: str
    address: str
    errors: List[str]
    response: List[Any]
