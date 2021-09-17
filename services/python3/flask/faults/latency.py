import time
from models import Fault


class Latency(Fault):
    def __init__(self, type, args):
        super().__init__(type, args)

    def run(self):
        time.sleep(self.args['delay'] / 1000)
