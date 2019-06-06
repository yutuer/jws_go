# -*- coding:utf-8 -*-
from __future__ import absolute_import
from __future__ import unicode_literals
from __future__ import print_function

import json
import uuid
import time
import gevent

from websocket import create_connection
import six

from locust import HttpLocust, TaskSet, task
from locust.events import request_success


class EchoTaskSet(TaskSet):
    def on_start(self):
        self.user_id = six.text_type(uuid.uuid4())
        ws = create_connection('ws://127.0.0.1:5000/echo')
        self.ws = ws

        def _receive():
            while True:
                res = ws.recv()
                data = json.loads(res)
                end_at = time.time()
                response_time = int((end_at - data['start_at']) * 1000000)
                request_success.fire(
                    request_type='WebSocket Recv',
                    name='test/ws/echo',
                    response_time=response_time,
                    response_length=len(res),
                )

        gevent.spawn(_receive)

    def on_quit(self):
        self.ws.close()

    @task
    def sent(self):
        start_at = time.time()
        body = json.dumps({'message': 'hello, world', 'user_id': self.user_id, 'start_at': start_at})
        self.ws.send(body)
        request_success.fire(
            request_type='WebSocket Sent',
            name='test/ws/echo',
            response_time=int((time.time() - start_at) * 1000000),
            response_length=len(body),
        )


class ChatTaskSet(TaskSet):
    def on_start(self):
        self.user_id = six.text_type(uuid.uuid4())
        ws = create_connection('ws://127.0.0.1:5000/chat')
        self.ws = ws

        def _receive():
            while True:
                res = ws.recv()
                data = json.loads(res)
                end_at = time.time()
                response_time = int((end_at - data['start_at']) * 1000000)
                request_success.fire(
                    request_type='WebSocket Recv',
                    name='test/ws/chat',
                    response_time=response_time,
                    response_length=len(res),
                )

        gevent.spawn(_receive)

    def on_quit(self):
        self.ws.close()

    @task
    def sent(self):
        start_at = time.time()
        body = json.dumps({'message': 'hello, world', 'user_id': self.user_id, 'start_at': start_at})
        self.ws.send(body)
        request_success.fire(
            request_type='WebSocket Sent',
            name='test/ws/chat',
            response_time=int((time.time() - start_at) * 1000000),
            response_length=len(body),
        )


class EchoLocust(HttpLocust):
    task_set = EchoTaskSet
    min_wait = 0
    max_wait = 100


class ChatLocust(HttpLocust):
    task_set = ChatTaskSet
    min_wait = 0
    max_wait = 100
