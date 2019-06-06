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

now_utc = datetime.datetime.utcnow()
local_tz = pytz.timezone('Asia/Shanghai')
now_utc = pytz.utc.localize(now_utc)
local_time = now_utc.astimezone(local_tz)
current_date = local_time.strftime('%Y%m%d')

islocal = False
localPath = "/Users/tq/bigdatatest/"
targetpay_path = 'uidcreattime-csv/pay/'
target_path = 'uidcreattime-csv'
firstDay = datetime.date(2016, 11, 10)

# firstDay = datetime.date(2017, 3, 7)
# target_path = '/Users/tq/bigdatatest/'
# targetpay_path = 'Users/tq/bigdatatest/pay/'  # 本地


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

def getDBDriver(tableName):
    return spark.read.format("jdbc").option("url", url).option("driver", "com.mysql.jdbc.Driver").option(
        "dbtable", tableName).option("user", properties['user']).option("password", properties['password']).load()


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
    json_buffer = StringIO.StringIO()
    pandasdf.to_json(json_buffer, orient='records')
    conn = boto.s3.connect_to_region(region)
    bucket = conn.get_bucket(bucket_name)
    full_key_name = s3_filepath
    k = bucket.new_key(full_key_name)
    k.set_contents_from_string(json_buffer.getvalue())
    return None


def upload_to_local(pandasdf, s3_filepath):
    #     for column in pandasdf.columns:
    #         for idx in pandasdf[column].index:
    #             x = pandasdf.get_value(idx, column)
    #             try:
    #                 x = unicode(x.encode('utf-8', 'ignore'), errors='ignore') if type(x) == unicode else unicode(str(x),errors='ignore')
    #                 pandasdf.set_value(idx, column, x)
    #             except Exception:
    #                 print
    #                 'encoding error: {0} {1}'.format(idx, column)
    #                 pandasdf.set_value(idx, column, '')
    #                 continue
    pandasdf.to_json(s3_filepath, orient='records')
    return None


def upload(pandasdf, region, bucket_name, s3_filepath):
    if islocal:
        upload_to_local(pandasdf, s3_filepath)
    else:
        upload_to_s3(pandasdf, region, bucket_name, s3_filepath)


properties = {
    "user": "test1",
    "password": "QmPhaQ8hYsxx"
}
uidTableName = "uid_creatTime"
acidTableName = "account_creatTime"
payTableName = "account_payTime"
url = "jdbc:mysql://54.223.192.252:3306/test"

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


def ltv_byGid_sqlalchemy_orm(df):
    # init_sqlalchemy()
    session = DBSession()
    x = []
    for dfdate in df.values:
        customer = LTVByGid()
        customer.gid = str(dfdate[0])
        customer.creat_time = str(dfdate[3])
        customer.people_count = str(dfdate[2])
        customer.days = str(dfdate[4])
        if dfdate[1] is None:
            customer.consume = float(0)
        else:
            customer.consume = float(dfdate[1]) / float(dfdate[2])
        x.append(customer)
    session.add_all(x)
    session.commit()


def ltv_bySid_sqlalchemy_orm(df):
    # init_sqlalchemy()
    session = DBSession()
    x = []
    for dfdate in df.values:
        customer = LTVBySid()
        customer.sid = str(dfdate[0])
        customer.creat_time = str(dfdate[3])
        customer.people_count = str(dfdate[2])
        customer.days = str(dfdate[4])
        if dfdate[1] is None:
            customer.consume = float(0)
        else:
            customer.consume = float(dfdate[1]) / float(dfdate[2])
        x.append(customer)
    session.add_all(x)
    session.commit()


def ltv_byCh_sqlalchemy_orm(df):
    # init_sqlalchemy()
    session = DBSession()
    x = []
    for dfdate in df.values:
        customer = LTVByCh()
        customer.ch = str(dfdate[0])
        customer.creat_time = str(dfdate[3])
        customer.people_count = str(dfdate[2])
        customer.days = str(dfdate[4])
        if dfdate[1] is None:
            customer.consume = float(0)
        else:
            customer.consume = float(dfdate[1]) / float(dfdate[2])
        x.append(customer)
    session.add_all(x)
    session.commit()


if __name__ == '__main__':
    begin = datetime.date(2016, 12, 11)
    end = datetime.date(2017, 1, 11)

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
        #读取历史上所有的支付log，内存占用率会很高可能
        df = spark.read.json(get_file_list(day, "pay/"))
        filter_df = df.where(df.utc8[:10] <= xx)
        print time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(time.time()))), "finish read original data"
        filter_df.createOrReplaceTempView('PayTable')
        preparetable = 'PayTable'


        spark.sql(u'''
                        SELECT
                            PayTable.userid AS uid,
                            SUM(PayTable.info.money) AS sumpay,
                            PayTable.gid AS gid
                        FROM PayTable
                        GROUP BY PayTable.gid,PayTable.userid
                        ''').createOrReplaceTempView('payByGid')

        spark.sql(u'''
                        SELECT
                            PayTable.accountid AS acid,
                            SUM(PayTable.info.money) AS sumpay,
                            PayTable.sid AS sid
                        FROM PayTable
                        GROUP BY PayTable.sid,PayTable.accountid
                        ''').createOrReplaceTempView('payBySid')

        spark.sql(u'''
                        SELECT
                            PayTable.userid AS uid,
                            SUM(PayTable.info.money) AS sumpay,
                            PayTable.channel AS ch
                        FROM PayTable
                        GROUP BY PayTable.channel,PayTable.userid
                        ''').createOrReplaceTempView('payByCh')

        # uid无Channel
        uidList = getTempDataList("uid_creatTime")
        df_old_uid = spark.read.json(uidList)
        final_uid_Result_fream = df_old_uid.where(df_old_uid.date <= xx)
        final_uid_Result_fream.createOrReplaceTempView('uidlog')
        # acid无channel
        acidList = getTempDataList("acid_creatTime")
        df_temp_acid = spark.read.json(acidList)
        final_acid_Result_fream = df_temp_acid.where(df_temp_acid.date <= xx)
        final_acid_Result_fream.createOrReplaceTempView('acidlog')
        # 渠道
        ChannelList = getTempDataList("uid_ByChannelcreatTime")
        df_temp_uidChannel = spark.read.json(ChannelList)
        final_uidChannel_Result_fream = df_temp_uidChannel.where(df_temp_uidChannel.date <= xx)
        final_uidChannel_Result_fream.createOrReplaceTempView('uidChannelLog')


        # 查出账号创建时间
        print time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(time.time()))), "start ltv"
        cToday = datetime.date(int(pp[0]), int(pp[1]), int(pp[2]))
        for i in range((cToday - firstDay).days + 1):
            creatDay = u''' "%s"''' % str(firstDay + datetime.timedelta(days=i))
            creatDayDateTime = firstDay + datetime.timedelta(days=i)
            theday = (cToday - creatDayDateTime).days + 1
            payResultByGid = spark.sql(u'''
                                            SELECT
                                                s.gid AS gid,
                                                SUM(t.pay) AS payNum,
                                                COUNT(s.uid) AS people,
                                                first(s.date) AS date
                                            FROM(
                                            (SELECT
                                                uid AS uid,
                                                date AS date ,
                                                gid AS gid
                                            FROM uidlog
                                            WHERE date = %s) s
                                            LEFT OUTER JOIN(
                                            SELECT
                                                uid AS uid,
                                                sumpay AS pay,
                                                gid AS gid
                                            FROM payByGid
                                            ) t
                                            on t.uid = s.uid)
                                            GROUP BY s.gid
                                        ''' % creatDay)

            finalLTV_ByGid = payResultByGid.withColumn("days", payResultByGid.people - payResultByGid.people + theday)
            payResultBySid = spark.sql(u'''
                                            SELECT
                                                s.sid As sid,
                                                SUM(t.payNum) AS payNum,
                                                COUNT(s.acid) AS people,
                                                first(s.date) AS date
                                            FROM(
                                            (SELECT
                                                acid AS acid,
                                                date AS date,
                                                sid AS sid
                                            FROM acidlog
                                            WHERE date = %s) s
                                            LEFT OUTER JOIN(
                                            SELECT
                                                acid AS acid,
                                                sumpay AS payNum,
                                                sid AS sid
                                            FROM payBySid
                                            ) t
                                            on t.acid = s.acid)
                                            GROUP BY s.sid
                                        ''' % creatDay)

            finalLTV_BySid = payResultBySid.withColumn("days", payResultBySid.people - payResultBySid.people + theday)
            payResultByCh = spark.sql(u'''
                                            SELECT
                                                s.ch AS ch,
                                                SUM(t.pay) AS payNum,
                                                COUNT(s.uid) AS people,
                                                first(s.date)
                                            FROM(
                                            (SELECT
                                                uid AS uid,
                                                date AS date ,
                                                ch AS ch
                                            FROM uidChannelLog
                                            WHERE date = %s) s
                                            LEFT OUTER JOIN(
                                            SELECT
                                                uid AS uid,
                                                sumpay AS pay,
                                                ch AS ch
                                            FROM payByCh
                                            ) t
                                            on t.uid = s.uid)
                                            GROUP BY s.ch
                                        ''' % creatDay)

            finalLTV_ByCh = payResultByCh.withColumn("days", payResultByCh.people - payResultByCh.people + theday)

            payResultByGid_pd = finalLTV_ByGid.toPandas()
            payResultBySid_pd = finalLTV_BySid.toPandas()
            payResultByCh_pd = finalLTV_ByCh.toPandas()

            # 替换 nan
            payResultByGid_pd = payResultByGid_pd.where(payResultByGid_pd.notnull(), None)
            payResultBySid_pd = payResultBySid_pd.where(payResultBySid_pd.notnull(), None)
            payResultByCh_pd = payResultByCh_pd.where(payResultByCh_pd.notnull(), None)
            print time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(time.time()))), "finish today ltv,begin ->sql"
            ltv_byGid_sqlalchemy_orm(payResultByGid_pd)
            ltv_bySid_sqlalchemy_orm(payResultBySid_pd)
            ltv_byCh_sqlalchemy_orm(payResultByCh_pd)
        print time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(time.time()))), "finish today data"

    spark.stop()