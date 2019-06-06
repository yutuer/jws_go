# -*- coding:utf-8 -*-
import pandas as pd
import codecs
import json
import io
import StringIO
from pandas.io.json import json_normalize
import datetime
from boto import s3
import os
import gc

def get_s3_tempdata(dataName):
    prefix = "uidcreattime-csv/"
    total_size = 0
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    ret = []
    if bucket:
        for k in bucket.list(prefix=prefix):
            if k.size <= 0:
                continue
            if dataName == "payLog":
                logsp = k.name.split('/')
                a = logsp[len(logsp) - 1]
                if dataName in a:
                    total_size += k.size
                    ret.append('s3://prodlog/' + k.name)
                    print('test :s3://prodlog/' + k.name, ''.join(a))
            else:
                logsp = k.name.split('/')
                a = logsp[len(logsp) - 1]
                csvName = a.split('.')[0]
                if csvName == dataName:
                    total_size += k.size
                    ret.append('s3://prodlog/' + k.name)
                    print('test :s3://prodlog/' + k.name, ''.join(a))
    print('total:%d' % (total_size / 1024.0 / 1024.0 / 1024.0))
    return ret[0]

def get_s3_list():
    prefix = "pay/"
    total_size = 0
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    ret = []
    if bucket:
        for k in bucket.list(prefix=prefix):
            if k.size <= 0:
                continue
            logsp = k.name.split('.')
            a = logsp[-4:-1]
            total_size += k.size
            ret.append('s3://prodlog/' + k.name)
            print('s3://prodlog/' + k.name, ''.join(a))
    print('total:%d' % (total_size / 1024.0 / 1024.0 / 1024.0))
    return ret


def search(path, word):
    name = []
    for filename in os.listdir(path):
        fp = os.path.join(path, filename)
        if os.path.isfile(fp) and word in filename:
            name.append(fp)
        elif os.path.isdir(fp):
            search(fp, word)
    print name
    return name


def readLogicPay(payPath):
    paylog = pd.read_json(payPath[0])
    paylog.rename(columns={'payNum': 'Money', 'date': 'PayTime'}, inplace=True)
    allPay = paylog[['Money', 'PayTime', 'uid', 'ch']]

    for x in range(len(payPath)):
        if x == 0:
            continue
        paylog = pd.read_json(payPath[x])
        paylog.rename(columns={'payNum': 'Money', 'date': 'PayTime'}, inplace=True)
        pay = paylog[['Money','PayTime','uid','ch']]
        allPay = allPay.append(pay)
        del pay
        gc.collect()

    return allPay


def getChTemp(uidLogPath):
    uidByCh = pd.read_json(uidLogPath)
    uid_temp = uidByCh[['uid', 'date', 'ch']]
    del uidByCh
    gc.collect()
    return uid_temp


def getUidByChPeopleCount(ch_temp):
    creatGroup = ch_temp.groupby([ch_temp['date'], ch_temp['ch']]).count()
    creatGroup.rename(columns={'uid': 'peopleCount'}, inplace=True)
    creatGroup.reset_index(inplace=True)
    return creatGroup


def getSumTemp(payMerge):
    data2 = payMerge.reset_index()
    data3 = data2.set_index(["ch", "date", "PayTime"])  # index is the new column created by reset_index
    running_sum = data3.groupby(level=[0, 1, 2]).sum().groupby(level=[0, 1]).cumsum()
    #     result = payMerge.groupby([payMerge['sid'], payMerge['date'], payMerge['PayTime']]).sum()
    del running_sum['index']
    temp_sum = running_sum.reset_index()
    return temp_sum


if __name__ == '__main__':
    #uidPathLog = pd.read_json("/data/s3-data/uid_creatTime.json")
    payPath = search("/data/s3-data/pay","pay")


    pay = readLogicPay(payPath)


    ch_temp = getChTemp("/data/s3-data/uid_ByChannelcreatTime.json")
    creatGroup = getUidByChPeopleCount(ch_temp)

    payMerge = pay.merge(ch_temp, left_on='ch', right_on='ch', how='left', suffixes=['', '_r'])
    del payMerge['ch_r']
    del pay
    gc.collect()

    temp_sum = getSumTemp(payMerge)
    finalpayMerge = pd.merge(temp_sum, creatGroup, on=['ch', 'date'], how='left', suffixes=['', '_r'])
    finalpayMerge['consume'] = finalpayMerge.Money / finalpayMerge.peopleCount
    finalpayMerge['day'] = (finalpayMerge['PayTime'].apply(lambda x: x.date())) - finalpayMerge['date'].apply(lambda x: x.date())
    finalpayMerge['day'] = finalpayMerge['day'].apply(lambda x: x.days)
    del finalpayMerge['PayTime']
    print finalpayMerge
    finalpayMerge.to_excel('ChLTV.xlsx',sheet_name='sheet1')
