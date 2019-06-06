# -*- coding: utf-8 -*-
from __future__ import absolute_import
from __future__ import unicode_literals
from __future__ import print_function

from collections import defaultdict
import json

from geventwebsocket.handler import WebSocketHandler
from gevent.pywsgi import WSGIServer
from flask import Flask, request
from werkzeug.exceptions import abort


app = Flask(__name__)

ctr = defaultdict(int)


@app.route('/echo')
def echo():
    ws = request.environ['wsgi.websocket']
    if not ws:
        abort(400)

    while True:
        message = ws.receive()
        if message is not None:
            r = json.loads(message)
            ctr[r['user_id']] += 1

        ws.send(message)


@app.route('/report')
def report():
    return '\n'.join(['{}:\t{}\n'.format(user_id, count) for user_id, count in ctr.items()])


socket_handlers = set()


@app.route('/chat')
def chat():
    ws = request.environ['wsgi.websocket']
    socket_handlers.add(ws)

    while True:
        message = ws.receive()
        for socket_handler in socket_handlers:
            try:
                socket_handler.send(message)
            except:
                socket_handlers.remove(socket_handler)


if __name__ == '__main__':
    http_server = WSGIServer(('', 5000), app, handler_class=WebSocketHandler)
    http_server.serve_forever()
