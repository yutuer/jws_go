# -*- coding: utf-8 -*-

import json
import os
import datetime
import xlwt
from boto import s3
import csv
import sys


def updateUid2CreatTime(userInfoDataPath):
    wordOne = "userinfo"

    dataUser = search(userInfoDataPath, wordOne)

    totaldata200 = {}
    totaldata201 = {}
    totaldata202 = {}
    totaldata203 = {}
    with open("./uid2Time_200.json") as f:
        data200 = json.load(f)
    with open("./uid2Time_201.json") as f:
        data201 = json.load(f)
    with open("./uid2Time_202.json") as f:
        data202 = json.load(f)
    with open("./uid2Time_203.json") as f:
        data203 = json.load(f)

    fileNum = 0
    for file in dataUser:
        fileNum += 1
        print "###", file, fileNum
        with open(file) as f:
            reader = csv.reader(f)
            for line in reader:
                if line[0] == "accountId":
                    rTime = line.index("注册时间")
                    continue
                try:
                    tmp = line[0].split(':')
                    uid = tmp[2]
                    if line[0][0:3] == '200':
                        if totaldata200.has_key(uid):
                            if line[rTime][0:10] < totaldata200[uid]:
                                totaldata200[uid] = line[rTime][0:10]
                        else:
                            totaldata200.setdefault(uid, line[rTime][0:10])

                    if line[0][0:3] == '201':
                        if totaldata201.has_key(uid):
                            if line[rTime][0:10] < totaldata201[uid]:
                                totaldata201[uid] = line[rTime][0:10]
                        else:
                            totaldata201.setdefault(uid, line[rTime][0:10])

                    if line[0][0:3] == '202':
                        if totaldata202.has_key(uid):
                            if line[rTime][0:10] < totaldata202[uid]:
                                totaldata202[uid] = line[rTime][0:10]
                        else:
                            totaldata202.setdefault(uid, line[rTime][0:10])

                    if line[0][0:3] == '203':
                        if totaldata203.has_key(uid):
                            if line[rTime][0:10] < totaldata203[uid]:
                                totaldata203[uid] = line[rTime][0:10]
                        else:
                            totaldata203.setdefault(uid, line[rTime][0:10])
                except Exception, e:
                    print Exception, e
                    print line[0], file

    json_str200 = json.dumps(data200)
    json_str201 = json.dumps(data201)
    json_str202 = json.dumps(data202)
    json_str203 = json.dumps(data203)
    store("uid2Time_200.json", json_str200)
    store("uid2Time_201.json", json_str201)
    store("uid2Time_202.json", json_str202)
    store("uid2Time_203.json", json_str203)

def store(jName,data):
    with open(jName, 'w') as json_file:
        json_file.write(data)

name = []
def search(path, word):
    for filename in os.listdir(path):
        fp = os.path.join(path, filename)
        if os.path.isfile(fp):
            name.append(fp)
        elif os.path.isdir(fp):
            search(fp, word)
    return name

if __name__ == '__main__':
    print sys.argv
    if len(sys.argv) < 2:
        print 'No action specified.'
        sys.exit()
    print sys.argv
    updateUid2CreatTime(sys.argv[1])