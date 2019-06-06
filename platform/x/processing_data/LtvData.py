# -*- coding: utf-8 -*-
import json
import os
import csv
import datetime
from boto import s3
import sys

def acid2uidCreatTIme():
    print "begin"

    ltvData_200 = {}
    ltvData_201 = {}
    ltvData_202 = {}
    ltvData_203 = {}


    with open("/home/ec2-user/acid2Time_200.json") as f:
        data200 = json.load(f)
    with open("/home/ec2-user/acid2Time_201.json") as f:
        data201 = json.load(f)
    with open("/home/ec2-user/acid2Time_202.json") as f:
        data202 = json.load(f)
    with open("/home/ec2-user/acid2Time_203.json") as f:
        data203 = json.load(f)

    print "read ok"


    for key in data200:
        tmp = key.split(':')
        uid = tmp[2]
        if uid not in ltvData_200.keys():
            ltvData_200.setdefault(uid,data200[key])
        elif ltvData_200[uid] > data200[key]:
            ltvData_200[uid] = data200[key]

    for key in data201:
        tmp = key.split(':')
        uid = tmp[2]
        if uid not in ltvData_201.keys():
            ltvData_201.setdefault(uid,data201[key])
        elif ltvData_201[uid] > data201[key]:
            ltvData_201[uid] = data201[key]

    for key in data202:
        tmp = key.split(':')
        uid = tmp[2]
        if uid not in ltvData_202.keys():
            ltvData_202.setdefault(uid,data202[key])
        elif ltvData_202[uid] > data202[key]:
            ltvData_202[uid] = data202[key]

    for key in data203:
        tmp = key.split(':')
        uid = tmp[2]
        if uid not in ltvData_203.keys():
            ltvData_203.setdefault(uid,data203[key])
        elif ltvData_203[uid] > data203[key]:
            ltvData_203[uid] = data203[key]

    json_str200 = json.dumps(ltvData_200)
    json_str201 = json.dumps(ltvData_201)
    json_str202 = json.dumps(ltvData_202)
    json_str203 = json.dumps(ltvData_203)

    store("uid2Time_200.json",json_str200)
    store("uid2Time_201.json",json_str201)
    store("uid2Time_202.json",json_str202)
    store("uid2Time_203.json",json_str203)


def store(jName,data):
    with open(jName, 'w') as json_file:
        json_file.write(data)


if __name__ == '__main__':
    acid2uidCreatTIme()