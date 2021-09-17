import threading
import time
from models import Fault
import gc
import logging

logger = logging.getLogger('service')


def create_leak(size, duration):
    logger.info(f"creating memory leak of {size}MB for {duration} seconds")
    leak = "a" * size * 1024 * 1024
    time.sleep(duration / 1000)
    del leak
    time.sleep(1)
    gc.collect()


class MemoryLeak(Fault):
    def __init__(self, type, args):
        super().__init__(type, args)

    def run(self):
        thread = threading.Thread(target=create_leak, args=(self.args['size'], self.args['duration']))
        thread.daemon = True  # Daemonize thread
        thread.start()  # Start the execution
