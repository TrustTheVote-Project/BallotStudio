#!/usr/bin/env python3

import threading
import time

now = time.time

class meta:
    def __init__(self, ttl=None):
        self.ttl = ttl

class Cache:
    def __init__(self):
        self.items = {}
        self.lock = threading.Lock()
        self.closer = threading.Condition(self.lock)
        self.t = threading.Thread(target=self.gcThread)
        self.t.daemon = True
        self.t.start()

    def close(self):
        with self.lock:
            self.closer.notify()
        self.t.join(1)

    def set(self, key, value, time=None):
        with self.lock:
            self.items[key] = (meta(now() + time), value)

    def get(self, key):
        with self.lock:
            mv = self.items.get(key)
            if mv:
                m, v = mv
                if (m.ttl is None) or (m.ttl > now()):
                    return v
            return None

    def gcThread(self):
        self.closer.acquire()
        while True:
            if self.closer.wait(15):
                # got close
                self.closer.release()
                return

            # do garbage collection
            t = now()
            todel = []
            for k,mv in self.items.items():
                m,v = mv
                if m.ttl and (m.ttl < t):
                    todel.append(k)
            for k in todel:
                self.items.pop(k)
