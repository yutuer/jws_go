# -*- coding: utf-8 -*-
import json
import os
import csv
import datetime
from boto import s3
import sys


def uid2CreatTime(userInfoDataPath):
    print "begin",datetime.datetime.now()
    wordOne = "userinfo"
    wordTwo = "charge"
    dataUser = search(userInfoDataPath, wordOne)
    #datacharge = search(chargeDataPath,wordTwo)
    totaldata200 = {}
    totaldata201 = {}
    totaldata202 = {}
    totaldata203 = {}

    fileNum = 0

    for file in dataUser:
        fileNum += 1
        print "###",file,fileNum
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
                            totaldata200.setdefault(uid,line[rTime][0:10])

                    if line[0][0:3] == '201':
                        if totaldata201.has_key(uid):
                            if line[rTime][0:10] < totaldata201[uid]:
                                totaldata201[uid] = line[rTime][0:10]
                        else:
                            totaldata201.setdefault(uid,line[rTime][0:10])

                    if line[0][0:3] == '202':
                        if totaldata202.has_key(uid):
                            if line[rTime][0:10] < totaldata202[uid]:
                                totaldata202[uid] = line[rTime][0:10]
                        else:
                            totaldata202.setdefault(uid,line[rTime][0:10])

                    if line[0][0:3] == '203':
                        if totaldata203.has_key(uid):
                            if line[rTime][0:10] < totaldata203[uid]:
                                totaldata203[uid] = line[rTime][0:10]
                        else:
                            totaldata203.setdefault(uid,line[rTime][0:10])
                except Exception,e:
                    print Exception,e
                    print line[0],file


    json_str200 = json.dumps(totaldata200)
    json_str201 = json.dumps(totaldata201)
    json_str202 = json.dumps(totaldata202)
    json_str203 = json.dumps(totaldata203)
    store("uid2Time_200.json",json_str200)
    store("uid2Time_201.json",json_str201)
    store("uid2Time_202.json",json_str202)
    store("uid2Time_203.json",json_str203)

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


def get_user_event_info(daterange):

    total_size = 0
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    retUser = []
    retPay = []
    if bucket:
        for k in bucket.list():
            if k.size <=0:
                continue
            logsp = k.name.split('_')
            a = logsp[len(logsp)-1].split('.')[0]
            #suffix = logsp[len(logsp)-1].split('.')[1]
            if len(logsp) < 3:
                continue
            if str(logsp[2]) == 'userinfo' and logsp[len(logsp)-1].split('.')[1] == 'csv':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retUser.append(k.name)
            if str(logsp[2]) == 'charge' and logsp[len(logsp)-1].split('.')[1] == 'log':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retPay.append(k.name)
    print('total:%d'%(total_size/1024.0/1024.0/1024.0))
    # print retUser
    # print retPay
    print "im here1"
    if not os.path.exists(daterange+ '/bilogs-csv'):
        print "im here2"
        os.makedirs(daterange+'/bilogs-csv')
    if not os.path.exists(daterange+'/charge'):
        print "im here3"
        os.makedirs(daterange+'/charge')
    payNum = 0
    userNum = 0
    while userNum < len(retUser):
        try:
            print "file Num",userNum
            downfile(retUser[userNum],daterange)
            print retUser[userNum],"File Down Succes"
            userNum += 1
        except Exception,e:
            print Exception,e
            print "except:",userNum
            print retUser[userNum],"File try again"
    while payNum < len(retPay):
        try:
            print "file Num",payNum
            downfile(retPay[payNum],daterange)
            print retPay[payNum],"File Down Succes"
            payNum += 1
        except Exception,e:
            print Exception,e
            print "except:",payNum
            print retPay[payNum],"File try again"

# key charge/2017-03-15/134_130134004019_charge_2017-03-15.log
# key bilogs-csv/2017-03-11/134_130134004013_userinfo_2017-03-11.csv
def downfile(s3key,dPath):
    print "im here4"
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    tmp = s3key.split('/')
    name = tmp[len(tmp)-1]
    print bucket
    if bucket:
        key = bucket.get_key(s3key)
        print "if bucket ",key
        if "userinfo" in s3key:
            key.get_contents_to_filename(dPath + '/bilogs-csv/' + name)
        else:
            key.get_contents_to_filename(dPath + '/charge/' + name)

attachments = []
def email_report(to, subject, text):
    from email.parser import Parser
    import requests

    key = 'key-02761952fa3330171f0b28cd3a502341'
    sandbox = 'mg.taiyouxi.cn'

    request_url = 'https://api.mailgun.net/v2/{0}/messages'.format(sandbox)
    request = requests.post(request_url, auth=('api', key),
                            files=attachments,
                            data={'from': 'data@'+sandbox,
                                'to': to,
                                'subject': subject,
                                'text': text})

    if 200 == request.status_code:
        sys.stdout.write('Body:   {0}\n'.format(request.text))
        return 0
    else:
        sys.stderr.write('Status: {0}\n'.format(request.status_code))
        sys.stderr.write('Body:   {0}\n'.format(request.text))
        return 1


if __name__ == '__main__':
    #get_user_event_info("/home/ec2-user/ltv")
    uid2CreatTime("/data/s3-data/userinfo")