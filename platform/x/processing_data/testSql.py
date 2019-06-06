# -*- coding:utf-8 -*-
from __future__ import print_function
#from pyspark import SparkContext
#from pyspark.sql import HiveContext
from pyspark.sql import SparkSession
import pyspark.sql.functions as func

import numpy as np
import pandas as pd
from pandas import ExcelWriter


import boto
from boto import s3
import StringIO

import pytz
import datetime

from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy import Column, Integer, String, create_engine
from sqlalchemy.orm import scoped_session, sessionmaker


from boto import s3
from pyspark.sql import SparkSession
from pyspark import SparkContext, SparkConf

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

def get_s3_list(daterange):
    prefix = "logics3/psid=4"
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


def min_time(x, y):
    if int(x.lt) > int(y.lt):
        return y
    else:
        return x





def upload_to_s3(pandasdf, region, bucket_name, s3_filepath):
    for column in pandasdf.columns:
        for idx in pandasdf[column].index:
            x = pandasdf.get_value(idx, column)
            try:
                x = unicode(x.encode('utf-8', 'ignore'), errors='ignore') if type(x) == unicode else unicode(str(x),
                                                                                                             errors='ignore')
                pandasdf.set_value(idx, column, x)
            except Exception:
                print
                'encoding error: {0} {1}'.format(idx, column)
                pandasdf.set_value(idx, column, '')
                continue
    csv_buffer = StringIO.StringIO()
    pandasdf.to_csv(csv_buffer)
    conn = boto.s3.connect_to_region(region)
    bucket = conn.get_bucket(bucket_name)
    full_key_name = s3_filepath
    k = bucket.new_key(full_key_name)
    k.set_contents_from_string(csv_buffer.getvalue())
    return None



Base = declarative_base()
dbname='mysql+mysqlconnector://test1:QmPhaQ8hYsxx@54.223.192.252:3306/test'
engine = create_engine(dbname, echo=False)
DBSession = scoped_session(sessionmaker(bind=engine))


class Customer(Base):
    __tablename__ = "bigdatatest33"
    id = Column(Integer, primary_key=True)
    date = Column(String(255))
    sid = Column(String(255))
    gid = Column(String(255))
    ch = Column(String(255))
    machine = Column(String(255))
    lt = Column(String(255))


# def init_sqlalchemy(dbname='mysql+mysqlconnector://test1:QmPhaQ8hYsxx@54.223.192.252:3306/test'):
#     global engine
#     engine = create_engine(dbname, echo=False)

def test_sqlalchemy_orm(df):
    # init_sqlalchemy()
    session = DBSession()
    x = []
    for dfdate in df.values:
        customer = Customer()
        customer.uid = dfdate[0].encode('utf-8')
        customer.sid = str(dfdate[3])
        customer.ch = dfdate[5].encode('utf-8')
        customer.date = str(dfdate[2])
        customer.lt = dfdate[1].encode('utf-8')
        customer.gid = str(dfdate[4])
        customer.machine = dfdate[6].encode('utf-8')
        print (customer.lt)
        x.append(customer)
    session.add_all(x)
    session.commit()

if __name__ == '__main__':
    now_utc = datetime.datetime.utcnow()
    local_tz = pytz.timezone('Asia/Shanghai')
    now_utc = pytz.utc.localize(now_utc)
    local_time = now_utc.astimezone(local_tz)
    current_date = local_time.strftime('%Y%m%d')
    target_path = 'uidcreattime-csv'
    target_path += '/%s' % current_date

    spark = SparkSession \
        .builder.enableHiveSupport() \
        .appName("Python Spark SQL basic example") \
        .config("spark.some.config.option", "some-value") \
        .getOrCreate()

    df = spark.read.json(get_s3_list("20161115"))

    df.createOrReplaceTempView('RawTable')
    preparetable = 'RawTable'

    result1 = spark.sql(u'''
            SELECT
                *
            FROM %s
            WHERE accountid is not null
            AND type_name!="CorpLevelChg"
            ''' % preparetable).cache().createOrReplaceTempView('jsonTable')

    # 获取当日log里的最早的一次登录
    spark.sql(u"""
            SELECT
                t.uid AS uid,
                SUBSTRING(t.lt,0,10) AS lt,
                CAST(jsonTable.utc8 as DATE) AS date,
                jsonTable.sid AS sid,
                jsonTable.gid AS gid,
                jsonTable.channel AS ch,
                jsonTable.info.MachineType AS machine
            FROM (
                SELECT
                    userid AS uid,
                    min(logtime) AS lt
                FROM jsonTable
                WHERE type_name="Login"
                GROUP BY userid
            ) t
            LEFT JOIN jsonTable ON jsonTable.userid = t.uid AND jsonTable.logtime = t.lt
            """).cache().createOrReplaceTempView('uid_ch')

    acidlog_df = spark.sql(u"""
    SELECT *
    FROM uid_ch
    """)
    false_channel = getFalesChannel()
    true_channel = getTrueChannel()
    # 替换channel
    uid_creatTimeLog = acidlog_df.replace(false_channel, true_channel, "ch")

    uid_creatTimeLog_pandas = uid_creatTimeLog.toPandas()
    test_sqlalchemy_orm(uid_creatTimeLog_pandas)
    upload_to_s3(uid_creatTimeLog_pandas,'cn-north-1', "prodlog", '/%s/test.csv'%target_path)
    spark.stop()


