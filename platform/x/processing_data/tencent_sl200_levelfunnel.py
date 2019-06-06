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
import sys, os
import pytz
import datetime

#email attachments
#attachments = [('attachment', (x, open(x))) for x in files]
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

#spark-submit --deploy-mode client --master yarn  tencent.py s3://prodlog/shard1001/ emrs3
#spark-submit --deploy-mode cluster --master yarn  tencent.py s3://prodlog/shard1001/ emrs3
now_utc = datetime.datetime.utcnow()
local_tz = pytz.timezone('Asia/Shanghai')
now_utc = pytz.utc.localize(now_utc)
local_time = now_utc.astimezone(local_tz)
# current_date = 'test/'+local_time.strftime('%Y%m%d')
current_date = local_time.strftime('%Y%m%d')
target_path = 'report'
target_path += '/%s'%current_date
if(len(sys.argv) > 3):
    target_path += '/%s'%sys.argv[3]


rawchannels = StringIO.StringIO(u'''渠道ID,渠道名称
130134024500,5Gwan
130134005600,当乐
130134000164,PPTV
130134043300,应用汇
130134002281,优酷
130134001500,UC
130134011000,安智
130134003800,豌豆荚
130134001300,360
130134004600,百度多酷
130134005400,小米
130134003234,拇指玩
130134001232,Vivo
130134001306,木蚂蚁
130134003210,悠悠村
130134015600,益玩
130134012700,OPPO
130134012000,华为
130134001272,金立
130134000201,4399
130134000159,搜狗
130134003900,联想
130134000144,应用宝
130134012600,PPS
130134000137,酷派
130134022300,新浪
130134000126,联通
130134000156,酷狗
130134015700,偶玩
130134015800,今日头条
130134312112,机锋
130134023200,咪咕
130134015200,魅族
130134001335,乐视
130134003257,迅雷
130134024900,美图
130134312120,猎宝
130134000001,英雄互娱
130134000135,三星
130134001258,暴风
130134312133,TT语音
130134000146,唱吧
130134015300,卓易
130134013400,虫虫
''')

channels = pd.read_csv(rawchannels, encoding='utf-8', usecols=[u'渠道ID', u'渠道名称'])
channels.rename(columns={u'渠道ID': 'ch', u'渠道名称': 'chn'}, inplace=True)


def as_attachments(name, content):
    global attachments
    attachments.append(('attachment', (name, content)))

def upload_stringio_to_s3(pandasdf, region, bucket_name, s3_filepath):
    conn = boto.s3.connect_to_region(region)
    bucket = conn.get_bucket(bucket_name)
    full_key_name = s3_filepath
    k = bucket.new_key(full_key_name)
    k.set_contents_from_string(pandasdf.getvalue())

    as_attachments(os.path.basename(s3_filepath), pandasdf.getvalue())
    return None

def upload_to_s3(pandasdf, region, bucket_name, s3_filepath):
    csv_buffer = StringIO.StringIO()
    pandasdf.to_csv(csv_buffer)
    conn = boto.s3.connect_to_region(region)
    bucket = conn.get_bucket(bucket_name)
    full_key_name = s3_filepath
    k = bucket.new_key(full_key_name)
    k.set_contents_from_string(csv_buffer.getvalue())

    as_attachments(os.path.basename(s3_filepath), csv_buffer.getvalue())
    return None

#logics3/psid=3055,20161229,20170110
def get_s3_list(daterange):
    prefix = daterange[0]
    total_size = 0
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    ret = []
    if bucket:
        for k in bucket.list(prefix=prefix):
            if k.size <=0:
                continue
            logsp = k.name.split('.')
            a = logsp[-4:-1]
            a.reverse()
            dt = ''.join(a)
            if dt <= daterange[2] and dt >=daterange[1]:
                total_size += k.size
                ret.append('s3://prodlog/'+k.name)
                print('s3://prodlog/'+k.name, ''.join(a))
    print('total:%d'%(total_size/1024.0/1024.0/1024.0))
    return ret

# accountid to userid
def prepare_cache(spark):
    # spark.table('allprepared')
    preparetable = "allprepared"
    if sys.argv[1] == "":
        pass
    else:
        #spark.sql(u'''CREATE TEMPORARY VIEW RawTable USING org.apache.spark.sql.json OPTIONS (path "%s")'''%sys.argv[1])
        alldf = spark.read.json(get_s3_list(sys.argv[1].split(',')))
        alldf.createOrReplaceTempView('RawTable')
        preparetable = 'RawTable'

    # SELECT * FROM RawTable WHERE accountid is not null AND dt >= 20160720
    spark.sql(" AND ".join([u'''
    SELECT
        *
    FROM %s
    WHERE accountid is not null'''%preparetable] + sys.argv[4:] )).cache().createOrReplaceTempView('jsonTable')

    spark.sql(u"""
    SELECT
        userid AS uid,
        min(sid) AS sid,
        min(gid) AS gid,
        first(channel) AS ch,
        first(info.MachineType) AS machine
    FROM jsonTable
    WHERE type_name="Login"
    GROUP BY userid
    """).cache().createOrReplaceTempView('account_ch')

    spark.sql(u"""
    SELECT
        t.uid AS uid,
        t.date AS date,
        account_ch.ch AS ch,
        account_ch.sid AS sid,
        account_ch.gid AS gid,
        account_ch.machine AS m
    FROM (
        SELECT
            userid AS uid,
            min(CAST(utc8 as DATE)) AS date
        FROM jsonTable
        WHERE type_name="CreateProfile"
        GROUP BY userid
    ) t
    LEFT JOIN account_ch ON account_ch.uid = t.uid
    """).cache().createOrReplaceTempView('creates_profiles')



    spark.sql(u"""
    SELECT
        userid AS uid,
        sid,
        gid,
        info.LoginTimes AS loginTimes,
        info.MachineType AS machine,
        channel AS ch,
        CAST(utc8 as DATE) AS date
    FROM jsonTable
    WHERE type_name="Login"
    """).cache().createOrReplaceTempView('loginsevents')


    spark.sql(u"""
    SELECT
        uid,
        min(sid) AS sid,
        min(gid) AS gid,
        first(machine) AS machine,
        min(date) AS date,
        first(ch) AS ch
    FROM loginsevents
    WHERE loginTimes = 1
    GROUP BY uid
    """).cache().createOrReplaceTempView('loginsAsCreateEvents')


'''
按照每个事件的独立帐号统计，看关卡流失
'''
def get_level_funnel(spark):
    df = spark.sql(u'''
    WITH events AS (
        SELECT
            info.TimeEvent AS event,
            userid AS uid
        FROM jsonTable
        WHERE type_name="ClientTimeEvent"
    )
    SELECT event,
        COUNT(DISTINCT t.uid) AS n
    FROM (
        SELECT
           creates.uid AS uid,
           first(creates.machine) AS machine,
           first(creates.date) AS date,
           SUM(DATEDIFF(future_activity.date,creates.date)) AS period
        FROM loginsAsCreateEvents AS creates
        LEFT JOIN loginsevents AS future_activity
        ON creates.uid = future_activity.uid
        GROUP BY creates.uid
    ) t
    LEFT JOIN events
    ON events.uid = t.uid
    --WHERE period=0
    GROUP BY event
    ORDER BY n DESC
    ''')
    return df.toPandas()

'''
按照每个事件的独立帐号统计，看关卡流失
按照帐号创建日期进行事件的切分，然后统计uid数量
creates_profiles
'''
def get_level_funnel_by_ch(spark):
    df = spark.sql(u'''
    WITH events AS (
        SELECT
            info.TimeEvent AS event,
            userid AS uid
        FROM jsonTable
        WHERE type_name="ClientTimeEvent"
    ), funnelevents AS (
        SELECT
            events.event AS event,
            events.uid AS uid,
            create_ch.ch AS ch,
            create_ch.date AS date
        FROM events
        RIGHT JOIN creates_profiles AS create_ch
        ON events.uid = create_ch.uid
    )
    SELECT
        COUNT(DISTINCT uid) AS events_accounts,
        ch, date, event
    FROM funnelevents
    GROUP BY ch, date, event
    ''')
    #ndf = df.groupBy('ch', 'date', 'event').agg(func.countDistinct(df.uid).alias('events_accounts'))
    # .groupBy('event').pivot('date').max('events_accounts')
    # 这种方式date会出问题，显示成数字
    return df.toPandas()





# def get_cohort_loginevent(spark):
#     df = spark.sql(u'''
#     ''')
#     return df.toPandas()

if __name__ == "__main__":
    spark = SparkSession\
        .builder.enableHiveSupport()\
        .appName("AWS")\
        .getOrCreate()

    bucketname = sys.argv[2]
    prepare_cache(spark)


    cohort_xls_out = StringIO.StringIO()
    cohort_xls = pd.ExcelWriter(cohort_xls_out, engine='xlsxwriter')

    #level funnel
    leveldf = get_level_funnel_by_ch(spark)
    ndf = pd.pivot_table(leveldf, values=['events_accounts'],
               index=['event'], columns=['ch', 'date'], aggfunc=np.sum)
    ndf2 = pd.pivot_table(leveldf, values=['events_accounts'],
               index=['event'], columns=['date'], aggfunc=np.sum)

    levelfunnel_xls_out = StringIO.StringIO()
    levelfunnel_xls = pd.ExcelWriter(levelfunnel_xls_out, engine='xlsxwriter')
    ndf.to_excel(levelfunnel_xls, sheet_name='level_ch')
    ndf2.to_excel(levelfunnel_xls, sheet_name="funnel_create")
    get_level_funnel(spark).to_excel(levelfunnel_xls, sheet_name="funnel_all")
    levelfunnel_xls.save()
    upload_stringio_to_s3(levelfunnel_xls_out, 'cn-north-1', bucketname, '/%s/level_funnel.xlsx'%target_path)

    spark.stop()

    #livedata@taiyouxi.cn
    sys.exit(email_report('livedata@taiyouxi.cn', u'IFSG 数据留存计算 %s'%(' '.join(sys.argv[3:])), u'''
    IFSG 数据留存计算
    测试数据效果
    %s
    '''%(' '.join(sys.argv))))
