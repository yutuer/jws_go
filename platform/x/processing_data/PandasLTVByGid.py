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
    paylog = pd.read_json(payPath)
    paylog.rename(columns={'payNum': 'Money', 'date': 'PayTime'}, inplace=True)
    allPay = paylog[['Money', 'PayTime', 'uid', 'gid']]

    # for x in range(len(payPath)):
    #     if x == 0:
    #         continue
    #     paylog = pd.read_json(payPath[x])
    #     paylog.rename(columns={'payNum': 'Money', 'date': 'PayTime'}, inplace=True)
    #     pay = paylog[['Money','PayTime','uid','gid']]
    #     allPay = allPay.append(pay)

    return allPay


def getUidTemp(uidLogPath):
    uidBygid = pd.read_json(uidLogPath)
    uid_temp = uidBygid[['uid', 'date', 'gid']]
    return uid_temp


def getUidByGidPeopleCount(uid_temp):
    creatGroup = uid_temp.groupby([uid_temp['date'], uid_temp['gid']]).count()
    creatGroup.rename(columns={'uid': 'peopleCount'}, inplace=True)
    creatGroup.reset_index(inplace=True)
    return creatGroup


def getSumTemp(payMerge):
    data2 = payMerge.reset_index()
    data3 = data2.set_index(["Gid","Date","PeopleCount", "PayTime"])  # index is the new column created by reset_index
    running_sum = data3.groupby(level=[0, 1, 2, 3]).sum().groupby(level=[0, 1]).cumsum()
    #     result = payMerge.groupby([payMerge['sid'], payMerge['date'], payMerge['PayTime']]).sum()
    del running_sum['index']
    temp_sum = running_sum.reset_index()
    return temp_sum


if __name__ == '__main__':
    summaryInfoList = []
    #uidPathLog = pd.read_json("/data/s3-data/uid_creatTime.json")
    payPath = search("/data/s3-data/pay","pay")

    uid_temp = getUidTemp("/data/s3-data/uid_creatTime.json")
    creatGroup = getUidByGidPeopleCount(uid_temp)
    for fileName in payPath:
        print fileName
        pay = readLogicPay(fileName)

        payMerge = pay.merge(uid_temp, left_on='uid', right_on='uid', how='left', suffixes=['', '_r'])
        finalpayMerge = pd.merge(payMerge, creatGroup, on=['gid', 'date'], how='left', suffixes=['', '_r'])
        del payMerge['gid_r']

        for x in finalpayMerge.iterrows():
            summaryInfo22 = {}
            info = x[1]
            summaryInfo22.setdefault("Gid", info.gid)
            summaryInfo22.setdefault("Date", info.date)
            summaryInfo22.setdefault("PayTime", info.PayTime)
            summaryInfo22.setdefault("PeopleCount", info.peopleCount)
            summaryInfo22.setdefault("Money", info.Money)
            summaryInfoList.append(summaryInfo22)

    from pandas.core.frame import DataFrame
    print "data row",len(summaryInfoList)
    oo = pd.DataFrame(summaryInfoList, columns=['Gid', 'Date', 'PeopleCount', 'PayTime', 'Money'])
    dd = oo.groupby(['Gid', 'Date', 'PeopleCount', 'PayTime']).sum()
    dd.to_excel('dd.xlsx', sheet_name='sheet1')
    dd.reset_index(inplace=True)
    print oo
    temp_sum = getSumTemp(dd)
    #finalpayMerge = pd.merge(temp_sum, creatGroup, on=['Gid', 'Date'], how='left', suffixes=['', '_r'])
    temp_sum['Consume'] = temp_sum.Money / temp_sum.PeopleCount
    temp_sum['Day'] = (temp_sum['PayTime'].apply(lambda x: x.date())) - temp_sum['Date'].apply(lambda x: x.date())
    temp_sum['Day'] = temp_sum['Day'].apply(lambda x: x.days)
    del temp_sum['PayTime']
    print temp_sum
    temp_sum.to_excel('GidLTV.xlsx',sheet_name='sheet1')
