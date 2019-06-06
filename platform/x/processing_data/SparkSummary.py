# -*- coding:utf-8 -*-
from pyspark import SparkContext, SparkConf
from pyspark.sql import SparkSession
from pyspark.sql import Row
from pyspark.sql.types import *
import boto
from boto import s3
import datetime
import pandas as pd
import StringIO
import pytz
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy import Column, Integer, String, create_engine, Float
from sqlalchemy.orm import scoped_session, sessionmaker
import os
import time

islocal = False
localPath = "/Users/tq/bigdatatest/"
targetpay_path = 'uidcreattime-csv/pay/'
target_path = 'uidcreattime-csv'
firstDay = datetime.date(2016, 11, 10)



def getFalesChannel():
    false_channel = [
        "130134201",
        "110134101106.0",
        "13013412700",
        "13013419400",
        "130134164",
        "47",
        "13013415800",
        "1301343257",
        "13013413400",
        "131",
        "142",
        "145",
        "180",
        "188",
        "196",
        "204",
        "210",
        "213",
        "215",
        "216",
        "223",
        "234",
        "250",
        "259",
        "262",
        "263",
        "265",
        "292",
        "296",
        "298",
        "306",
        "309",
        "313",
        "1301341500",
        "1301341232",
        "13013416000",
        "13013411000",
        "1301344600",
        "1301345600",
        "13013412000",
        "1301341272",
        "13013415900",
        "130134137",
        "1301371335",
        "130134126",
        "1301343900",
        "13013415200",
        "13013423200",
        "1301343234",
        "130134312108",
        "1301341300",
        "130134312173",
        "5000",
        "130134159",
        "1301343800",
        "1301345400",
        "1301341",
        "130134144",
        "13013443300",
        "268",
        "350",
        "389"
    ]
    return false_channel


def getTrueChannel():
    true_channel = [
        "130134000201",
        "110134101106",
        "130134012700",
        "130134019400",
        "130134000164",
        "47",
        "130134015800",
        "130134003257",
        "130134013400",
        "131",
        "142",
        "145",
        "180",
        "188",
        "196",
        "204",
        "210",
        "213",
        "215",
        "216",
        "223",
        "234",
        "250",
        "259",
        "262",
        "263",
        "265",
        "292",
        "296",
        "298",
        "306",
        "309",
        "313",
        "130134001500",
        "130134001232",
        "130134016000",
        "130134011000",
        "130134004600",
        "130134005600",
        "130134012000",
        "130134001272",
        "130134015900",
        "130134000137",
        "130134001335",
        "130134000126",
        "130134003900",
        "130134015200",
        "130134023200",
        "130134003234",
        "130134312108",
        "130134001300",
        "130134312173",
        "5000",
        "130134000159",
        "130134003800",
        "130134005400",
        "130134000001",
        "130134000144",
        "130134043300",
        "268",
        "350",
        "389"
    ]
    return true_channel

def getQuickChannel():
    quickChannel =[
        "27",
        "106",
        "23",
        "33",
        "5",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "8888",
        "9",
        "20217",
        "21",
        "10",
        "14",
        "4",
        "24",
        "26",
        "68",
        "43",
        "75",
        "53",
        "29",
        "70",
        "69",
        "16",
        "140",
        "12",
        "332",
        "17",
        "28",
        "11",
        "15",
        "85",
        "32",
        "6",
        "8888",
        "8888",
        "8888"
    ]
    return quickChannel

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
            if dt == daterange:
                total_size += k.size
                ret.append('s3://prodlog/' + k.name)
                print('s3://prodlog/' + k.name, ''.join(a))
    print('total:%d' % (total_size / 1024.0 / 1024.0 / 1024.0))
    return ret


def get_local_list(daterange):
    return "/Users/tq/bigdatatest/%s.log" % daterange


def get_file_list(daterange, prefixpath):
    if islocal:

        return get_local_list(daterange)
    else:
        return get_s3_list(daterange, prefixpath)


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


def getTempDataList(dataName):
    if islocal:
        return search(localPath, dataName)
    else:
        return get_s3_tempdata(dataName)


def search(path, word):
    name = []
    if word == "payLog":
        for filename in os.listdir("/Users/tq/bigdatatest/pay"):
            fp = os.path.join("/Users/tq/bigdatatest/pay", filename)
            if word in filename:
                name.append(fp)
    else:
        for filename in os.listdir(path):
            fp = os.path.join(path, filename)
            if os.path.isfile(fp) and word in filename:
                name.append(fp)
    print name
    return name

Base = declarative_base()
dbname = 'mysql+mysqlconnector://test1:QmPhaQ8hYsxx@54.223.192.252:3306/test'
engine = create_engine(dbname, echo=False)
DBSession = scoped_session(sessionmaker(bind=engine))

class LTVByGid(Base):
    __tablename__ = "ltv_byGid"
    id = Column(Integer, primary_key=True)
    gid = Column(String(255))
    creat_time = Column(String(255))
    people_count = Column(String(255))
    days = Column(String(255))
    consume = Column(Float, nullable=True)


class LTVBySid(Base):
    __tablename__ = "ltv_bySid"
    id = Column(Integer, primary_key=True)
    sid = Column(String(255))
    creat_time = Column(String(255))
    people_count = Column(String(255))
    days = Column(String(255))
    consume = Column(Float, nullable=True)


class LTVByCh(Base):
    __tablename__ = "ltv_byCh"
    id = Column(Integer, primary_key=True)
    ch = Column(String(255))
    creat_time = Column(String(255))
    people_count = Column(String(255))
    days = Column(String(255))
    consume = Column(Float, nullable=True)



if __name__ == '__main__':
    begin = datetime.date(2017, 2, 1)
    end = datetime.date(2017, 6, 21)
    quick_channel = getQuickChannel()
    true_channel = getTrueChannel()
    fales_channel = getFalesChannel()
    print time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(time.time()))), "creat spark session"
    spark = SparkSession \
        .builder.enableHiveSupport() \
        .appName("Python Spark SQL basic example") \
        .config("spark.some.config.option", "some-value") \
        .getOrCreate()

    for i in range((end - begin).days + 1):
        xx = str(begin + datetime.timedelta(days=i))
        pp = xx.split('-')
        day = "".join(pp)
        print time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(time.time()))), "start:", day

        df = spark.read.json(get_file_list(day, "dataForSpark/"))
        print time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(time.time()))), "finish read original data"
        df.createOrReplaceTempView('RawTable')
        preparetable = 'RawTable'

        spark.sql(u'''
            SELECT
                *
            FROM %s
            WHERE accountid is not null
            AND type_name!="CorpLevelChg"
            ''' % preparetable).createOrReplaceTempView('jsonTable')
        print time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(time.time()))), "finish creat today jsonTable"

        # 计算当日登录
        login_df = spark.sql(u'''
                        SELECT
                                jsonTable.userid as uid,
                                jsonTable.accountid AS acid,
                                jsonTable.gid AS gid,
                                jsonTable.sid AS sid,
                                jsonTable.channel AS ch
                        FROM jsonTable
                        WHERE type_name="Login"
            ''')
        # 错误渠道号-->正确的渠道号-->Quick渠道号
        new_channel_ = login_df.replace(fales_channel, true_channel, "ch")
        new_channel_quick = new_channel_.replace(true_channel, quick_channel, "ch")

        # 计算当日分大区登录人数
        login_gid_df = login_df.groupBy(['gid']).count()
        login_gid = login_gid_df.withColumnRenamed("count", "loginPlayer")
        # print login_df.toPandas()
        # 计算当日分服务器登录人数
        login_sid_df = login_df.groupBy(['sid']).count()
        login_sid = login_sid_df.withColumnRenamed("count", "loginPlayer")
        # 计算当日分渠道登录人数
        login_ch_df = new_channel_quick.groupBy(['ch']).count()
        login_ch = login_ch_df.withColumnRenamed("count", "loginPlayer")

        # 计算新增 读creatTime系列json文件 在S3
        # 计算分大区新增
        uidList = getTempDataList("uid_creatTime")
        df_temp_uid = spark.read.json(uidList)
        new_uid = df_temp_uid.where(df_temp_uid.date == xx).groupBy(['gid']).count()
        new_uid_name = new_uid.withColumnRenamed("count", "newPlayer")
        # 计算分服务器新增
        acidList = getTempDataList("acid_creatTime")
        df_temp_acid = spark.read.json(acidList)
        new_acid = df_temp_acid.where(df_temp_acid.date == xx).groupBy(['sid']).count()
        new_acid_name = new_acid.withColumnRenamed("count", "newPlayer")
        # 计算分渠道新增
        ChannelList = getTempDataList("uid_ByChannelcreatTime")
        df_temp_uidChannel = spark.read.json(ChannelList)
        new_channelUid = df_temp_uidChannel.where(df_temp_uidChannel.date == xx).groupBy(['ch']).count()
        new_channel_name_ = new_channelUid.withColumnRenamed("count", "newPlayer")
        # 将渠道号 -->Quick渠道号
        new_channel_name = new_channel_name_.replace(true_channel, quick_channel, "ch")


        paylog_data = login_ch.replace(true_channel, quick_channel, "ch")

        # 读取支付log
        pay_temp_df = spark.read.json(get_file_list(day, "pay/"))
        pay_temp = pay_temp_df.where(pay_temp.utc8[:10] == '2016-11-10')
        pay_temp.createOrReplaceTempView('PayTable')

        payBySid = spark.sql(u"""
                        SELECT PayTable.sid AS sid,
                               SUM(PayTable.info.money) AS sumpay,
                               COUNT(*) AS PayNum,
                               COUNT(DISTINCT PayTable.userid) AS PayPeoples
                        FROM PayTable
                        GROUP BY PayTable.sid
            """)
        payByGid = spark.sql(u"""
                        SELECT PayTable.gid AS gid,
                               SUM(PayTable.info.money) AS sumpay,
                               COUNT(*) AS PayNum,
                               COUNT(DISTINCT PayTable.userid) AS PayPeoples
                        FROM PayTable
                        GROUP BY PayTable.gid
            """)
        payByChannel = spark.sql(u"""
                        SELECT PayTable.channel AS ch,
                               SUM(PayTable.info.money) AS sumpay,
                               COUNT(*) AS PayNum,
                               COUNT(DISTINCT PayTable.userid) AS PayPeoples
                        FROM PayTable
                        GROUP BY PayTable.channel
            """)

        # 开始分别依据，gid，sid，ch为key来合并数据
        gid_join = login_gid.join(new_uid_name, "gid", 'outer').join(payByGid, "gid", 'outer')

        sid_join = login_sid.join(new_sid_name, "sid", 'outer').join(payBySid, "sid", 'outer')

        ch_join = login_ch.join(new_channel_name, "ch", 'outer').join(payByChannel, "ch", 'outer')
        # 开始计算 付费率，arpu，arrpu
        # 分服务器
        s_sid = sid_join.withColumn("payPercent", sid_join.PayPeoples / sid_join.loginPlayer)
        arpu_sid = s_sid.withColumn("ARPU", s_sid.sumpay / s_sid.loginPlayer)
        summory_sid = arpu_sid.withColumn("ARPPU", arpu_sid.sumpay / arpu_sid.PayPeoples)
        # 分大区
        s_gid = gid_join.withColumn("payPercent", gid_join.PayPeoples / gid_join.loginPlayer)
        arpu_gid = s_gid.withColumn("ARPU", s_gid.sumpay / s_gid.loginPlayer)
        summory_gid = arpu_gid.withColumn("ARPPU", arpu_gid.sumpay / arpu_gid.PayPeoples)
        # 分渠道
        s_ch = ch_join.withColumn("payPercent", ch_join.PayPeoples / ch_join.loginPlayer)
        arpu_ch = s_ch.withColumn("ARPU", s_ch.sumpay / s_ch.loginPlayer)
        summory_ch = arpu_ch.withColumn("ARPPU", arpu_ch.sumpay / arpu_ch.PayPeoples)





















