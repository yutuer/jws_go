# -*- coding: utf-8 -*-
import datetime
import json
from xlrd import open_workbook
import MySQLdb.cursors
import MySQLdb  # 载入连接数据库模块
from boto import s3
import os
import time
def test():
    a = [1,2,6,3,4]
    b = [20,21,22,23,24]
    try:
        for x in range(len(a)):
            print b[a[x]]

    except Exception,e:
        print Exception,e
        print

def getRelativedate(date):
    x = date.split('-')
    if x[1].startswith("0"):
        mo = x[1][1]
    else:
        mo = x[1]
    if x[2].startswith("0"):
        day = x[2][1]
    else:
        day = x[2]

    return [x[0],mo,day]
def getRelativedays(dateStart,dateEnd):
    s = getRelativedate(dateStart)
    e = getRelativedate(dateEnd)
    print s[0], s[1], s[2]
    print e[0], e[1], e[2]
    d1 = datetime.date(int(s[0]), int(s[1]), int(s[2]))
    d2 = datetime.date(int(e[0]), int(e[1]), int(e[2]))
    print ((d2 - d1).days)

def test():
    with open("/Users/tq/Downloads/logics_shard1009.07.03.2017.log") as f:
        for line in f:
            d = json.loads(line)
            if d['type_name'] == 'AddGuildInventory':
                print d
def testexcel():
    wb1 = open_workbook("/Users/tq/Documents/diffExcel/after/FTE.xlsx" )
    x = wb1.sheet_by_name('llll')
    nr1 = x.nrows
    print nr1

def testDate():
    begin = datetime.date(2017,2,1)
    end = datetime.date(2017,3,7)
    for i in range((end - begin).days+1):
        day = begin + datetime.timedelta(days=i)
        print "I:",i
        print str(day)

def getTime():
    # ! /usr/bin/python
    # filename  conn.py
    server = {}
    try:  # 尝试连接数据库
        conn = MySQLdb.connect(host="54.223.192.252", user="test1", passwd="QmPhaQ8hYsxx", db="test")
    except MySQLdb.OperationalError, message:  # 连接失败提示
        print "link error"

    cursor = conn.cursor()  # 定义连接对象
    cursor.execute("select sid,start_time from server_startTime")  # 使用cursor提供的方法来执行查询语句
    data = cursor.fetchall()  # 使用fetchall方法返回所有查询结果
    for x in data:
        server.setdefault(x[0],x[1])
    print data  # 打印查询结果
    print server
    print server['4003']
    cursor.close()  # 关闭cursor对象
    conn.close()  # 关闭数据库链接

def get_s3_list(daterange, prefixpath):
    prefix = prefixpath
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
            a.reverse()
            dt = ''.join(a)
            if dt <= daterange:
                total_size += k.size
                ret.append('s3://prodlog/' + k.name)
                print('s3://prodlog/' + k.name, ''.join(a))
    print('total:%d' % (total_size / 1024.0 / 1024.0 / 1024.0))
    return ret

def get_s3_list22(daterange):
    prefix = "logics3/"
    total_size = 0
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    ret = []
    if bucket:
        for k in bucket.list(prefix=prefix):
            if k.size <= 0:
                continue
            logsp = k.name.split('/')
            #print logsp
            # a = logsp[-4:-1]
            # a.reverse()
            # dt = ''.join(a)
            # print dt
            dt = logsp[1]
            if dt == daterange:
                total_size += k.size
                ret.append(logsp[2])
                #print('s3://prodlog/' + k.name, ''.join(logsp))
    print('total:%d' % (total_size / 1024.0 / 1024.0 / 1024.0))
    print len(ret)
    print ret
    return ret

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
    return ret

def search1(path, word):
    name = []
    for filename in os.listdir(path):
        fp = os.path.join(path, filename)
        if os.path.isfile(fp) and word in filename:
            name.append(fp)
        elif os.path.isdir(fp):
            search(fp, word)
    print name
    return name


def search(path, word):
    name = []
    if word == "payLog":
        for filename in os.listdir("/Users/tq/bigdatatest/pay"):
            fp = os.path.join(path, filename)
            print fp
            print word in filename
            if word in filename:
                name.append(fp)
    else:
        for filename in os.listdir(path):
            fp = os.path.join(path, filename)
            if os.path.isfile(fp) and word in filename:
                name.append(fp)
    print "hehehehe",name
    return name


def get_s3_list333(daterange, prefixpath):
    prefix = prefixpath
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
            a.reverse()
            dt = ''.join(a)
            if dt == daterange:
                total_size += k.size
                ret.append('s3://prodlog/' + k.name)
                print('s3://prodlog/' + k.name, ''.join(a))
    print('total:%d' % (total_size / 1024.0 / 1024.0 / 1024.0))
    return ret

def paylogcc(path):
    countPay = 0
    for filename in os.listdir(path):
        fp = os.path.join(path, filename)
        with open(fp) as f:
            for line in f:
                try:
                    dataLine = json.loads(line,encoding="utf-8")
                    if dataLine['type_name'] == "IAP":
                        countPay += dataLine['Money']
                except Exception,e:
                    print dataLine


def getTheDay(dayBegin,dayEnd):
    b = dayBegin.split('-')
    e = dayEnd.split('-')
    bDay = datetime.date(int(b[0]),int(b[1]),int(b[2]))
    eDay = datetime.date(int(e[0]),int(e[1]),int(e[2]))
    print (eDay-bDay).days+1
    return (eDay-bDay).days+1

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

if __name__ == '__main__':
    #getRelativedays("2017-11-10","2017-12-12")
    #getTime()
    #testDate()
    # 转换成localtime
    #print time.localtime(int(time.time()))
    # 转换成新的时间格式(2016-05-09 18:59:20)
    #print time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(time.time())))

    #get_s3_tempdata("payLog")
    #search("/Users/tq/bigdatatest/pay/20170621","payLog")
    # f = "s3://prodlog/logics3/psid=4004/logics_shard4004.11.11.2016.gz"
    # Gz = f.split("/")
    # newS3Name = "/".join(Gz[-3:])
    # ss = newS3Name.split(".")
    # print ".".join(ss[:len(ss)-1])
    #get_s3_list("20161111","pay/")
    # b = get_s3_list22("psid=1001")
    # print list(set(b).difference(set(a)))
    #get_s3_list333("20161111", "dataForSpark/")
    #get_s3_tempdata('uidChannelLog')
    import MySQLdb

    database = MySQLdb.connect(host="54.223.192.252", user="test1", passwd="QmPhaQ8hYsxx", db="test")
    cursor = database.cursor()
    date1 = cursor.execute('''select gid,creat_time from ltv_byGid ''')
    # data1 = cursor.fetchall()
    print type(date1)
    date = cursor.execute(
        '''select * from ltv_byGid where creat_time = "2016-11-10" and days = (select max(days) from ltv_byGid where creat_time = "2016-11-10") ''')
    print date