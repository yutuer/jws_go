# -*- coding:utf-8 -*-
from __future__ import absolute_import
from __future__ import unicode_literals
from __future__ import print_function

import json
import uuid
import time
import gevent
import logging
import random, string
from gevent import GreenletExit
from gevent.subprocess import Popen, PIPE

logger = logging.getLogger(__name__)
from websocket import create_connection, WebSocketException
import six

from locust import HttpLocust, TaskSet, task, events
from locust.exception import StopLocust
from locust.events import request_success

botx_subp = None
def bot_start_worker():
    global botx_subp
    sub = Popen(['/Users/timesking/Projects/PlanX/ServerX/src/vcs.taiyouxi.net/botx/botx mbot -s 0 replay'], stdout=PIPE, shell=True)
    botx_subp = sub
    # for l in sub.stdout.readline():
    #     logger.info(l)
    logger.info("bot start")
    sub.communicate()
    # out, err = sub.communicate()
    # logger.info(out)

def bot_start():
    gevent.spawn(bot_start_worker)
    logger.info("start bot worker")
    time.sleep(3)
    logger.info("start bot worker good")

def bot_stop():
    global botx_subp
    if botx_subp:
        botx_subp.terminate()
        botx_subp = None
        logger.info("bot stop")


# events.locust_stop_hatching += bot_stop
# events.locust_start_hatching += bot_start

class GamexBotTaskSet(TaskSet):
    def run(self, *args, **kwargs):
            try:
                super(GamexBotTaskSet, self).run(args, kwargs)
            except (GreenletExit, StopLocust):
                if hasattr(self, "on_quit"):
                    self.on_quit()
                raise

    def on_start(self):
        pass
        # self.user_id = six.text_type(uuid.uuid4())
        ws = create_connection('ws://127.0.0.1:9080/ws')
        self.ws = ws
        self.run = True


        def _receive(p):
            while p.run:
                try:
                    res = ws.recv()
                    data = json.loads(res)
                    if data['Type'] == 'quit':
                        p.run = False
                    # end_at = time.time()
                    # response_time = int((end_at - data['start_at']) * 1000000)
                    request_success.fire(
                        request_type=data['Type'],
                        name=data['Name'],
                        # response_time=response_time,
                        response_time=1,
                        # response_length=len(res),
                        response_length=1,
                    )
                except:
                    logger.info("ws recv error")
                    break


        gevent.spawn(_receive, self)

    def on_quit(self):
        logger.info("on_quit")
        self.run = False
        self.ws.close()


    @task
    def sent(self):
        if not self.run:
            raise StopLocust
        pass

class GamexLocust(HttpLocust):
    task_set = GamexBotTaskSet
    min_wait = 1000
    max_wait = 5000
