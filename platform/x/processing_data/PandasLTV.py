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


if __name__ == '__main__':
    #acidURL = get_s3_tempdata("acid_creatTime")

    #payURLs = get_s3_list()

    acidBysid = pd.read_json("/data/s3-data/acid_creatTime.json")
    payPath = search("/data/s3-data/pay","pay")


    paylog = pd.read_json(payPath[0])
    paylog.rename(columns={'payNum': 'Money', 'date': 'PayTime'}, inplace=True)
    allPay = paylog[['Money', 'PayTime', 'acid', 'sid']]

    for x in range(len(payPath)):
        if x == 0:
            continue
        paylog = pd.read_json(payPath[x])
        paylog.rename(columns={'payNum': 'Money', 'date': 'PayTime'}, inplace=True)
        pay = paylog[['Money','PayTime','acid','sid']]
        allPay = allPay.append(pay)

    acid_temp = acidBysid[['acid', 'date', 'sid']]



    creatGroup = acid_temp.groupby([acid_temp['date'], acid_temp['sid']]).count()
    creatGroup.rename(columns={'acid': 'peopleCount'}, inplace=True)
    creatGroup.reset_index(inplace=True)
    payMerge = allPay.merge(acid_temp, left_on='acid', right_on='acid', how='left',suffixes=['', '_r'])
    del payMerge['sid_r']

    data2 = payMerge.reset_index()
    data3 = data2.set_index(["sid", "date", "PayTime"])  # index is the new column created by reset_index
    running_sum = data3.groupby(level=[0, 1, 2]).sum().groupby(level=[0, 1]).cumsum()
    #     result = payMerge.groupby([payMerge['sid'], payMerge['date'], payMerge['PayTime']]).sum()
    del running_sum['index']
    temp_sum = running_sum.reset_index()

    finalpayMerge = pd.merge(temp_sum, creatGroup, on=['sid', 'date'], how='left', suffixes=['', '_r'])

    finalpayMerge['consume'] = finalpayMerge.Money / finalpayMerge.peopleCount
    finalpayMerge['day'] = (finalpayMerge['PayTime'].apply(lambda x: x.date())) - finalpayMerge['date'].apply(lambda x: x.date())
    finalpayMerge['day'] = finalpayMerge['day'].apply(lambda x: x.days)

    del finalpayMerge['PayTime']

    print finalpayMerge
    finalpayMerge.to_excel('fool.xlsx',sheet_name='sheet1')

    # with open('/Users/tq/Downloads/20161110_paylog.json', 'r') as json_data:
    #     json_buffer = StringIO.StringIO()
    #     for line in json_data:
    #         tempMap = {}
    #         json_dicts = json.loads(line)  # 读取json数据为list[dict]结构
    #         tempMap.setdefault('sid', json_dicts['sid'])
    #         tempMap.setdefault('acid', json_dicts['accountid'])
    #         tempMap.setdefault('Money', json_dicts['info']['Money'])
    #         tempMap.setdefault('PayTime', json_dicts['utc8'][0:10])
    #         json_temp = json.dumps(tempMap)
    #         json_buffer.writelines(json_temp + "\n")
    #
    # pay = pd.read_json(json_buffer.getvalue(), lines=True)
