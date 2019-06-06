#!/bin/bash
#httpie

http --json POST 127.0.0.1:8791/api/v2/match/boss1 id=1:0:afdfds1 h:=true l:=30
http --json POST 127.0.0.1:8791/api/v2/match/boss1 id=1:0:afdfds2 h:=true l:=30


http --json POST 127.0.0.1:8791/api/v2/match/boss1?cancel=1 id=1:0:afdfds2 h:=true l:=30

http --json POST 127.0.0.1:8791/api/v2/match/boss1 id=1:0:afdfds3 h:=true l:=30

sleep 10
http --json POST 127.0.0.1:8791/api/v2/match/boss1 id=1:0:afdfds4 h:=true l:=30



