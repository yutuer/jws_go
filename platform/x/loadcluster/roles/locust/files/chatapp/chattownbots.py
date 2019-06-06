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

logger = logging.getLogger(__name__)
from websocket import create_connection, WebSocketException
import six

from locust import HttpLocust, TaskSet, task, events
from locust.events import request_success


def randomword(length):
   return ''.join(random.choice(string.lowercase) for i in range(length))

class EchoTaskSet(TaskSet):

    def run(self, *args, **kwargs):
            try:
                super(EchoTaskSet, self).run(args, kwargs)
            except GreenletExit:
                if hasattr(self, "on_quit"):
                    self.on_quit()
                raise

    def on_start(self):
        self.user_id = six.text_type(uuid.uuid4())
        self.create_new_ws(False)

        def _receive(p):
            while True:
                if p.ws is None:
                    logger.info("ws reconnect happen")
                    p.create_new_ws(True)

                try:
                    res = p.ws.recv()
                except WebSocketException:
                    p.ws = None
                except:
                    p.ws = None
                if res.startswith('msg'):
                    jres = res.split('msg:')[1]
                    data = json.loads(jres)
                    end_at = time.time()
                    response_time = int((end_at - data['start_at']) * 1000000)
                    request_success.fire(
                        request_type='WebSocket Recv',
                        name='Recv Say',
                        response_time=response_time,
                        response_length=len(jres),
                    )
                elif res.startswith('Room') or res.startswith('UnReg'):
                    request_success.fire(
                        request_type='WebSocket Recv',
                        name='Recv Someone Joinin/UnReg',
                        response_time=1,
                        response_length=1,
                    )
                else:
                    request_success.fire(
                        request_type='WebSocket Recv',
                        name='Recv Unknown',
                        response_time=1,
                        response_length=1,
                    )

        gevent.spawn(_receive, self)

    def on_quit(self):
        self.ws.close()

    def create_new_ws(self, reconnect):
        ws = create_connection('%s/ws'%self.client.base_url)
        self.ws = ws
        self.reg(reconnect)

    def reg(self, reconnect):
        try:
            self.ws.send('Reg:loadroom:%s'%self.user_id)
            regname = 'Sent Reg'
            if reconnect is True:
                regname = 'Sent Retry Reg'

            request_success.fire(
                request_type='WebSocket Sent',
                name=regname,
                response_time=1,
                response_length=1,
            )
        except WebSocketException:
            self.ws = None
        except:
            p.ws = None

    @task
    def sent(self):
        try:
            start_at = time.time()
            body = json.dumps({'message': randomword(255), 'user_id': self.user_id, 'start_at': start_at})
            msg = 'Say:loadroom:msg:%s'%body
            self.ws.send(msg)
            request_success.fire(
                request_type='WebSocket Sent',
                name='Sent Say',
                response_time=int((time.time() - start_at) * 1000000),
                response_length=len(body),
            )
        except WebSocketException:
            self.ws = None
        except:
            self.ws = None


class EchoLocust(HttpLocust):
    task_set = EchoTaskSet
    min_wait = 1000
    max_wait = 5000
