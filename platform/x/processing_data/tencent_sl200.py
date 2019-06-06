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

#logics3/psid=3
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
        WHERE type_name="FirstPay"
        GROUP BY userid
    ) t
    LEFT JOIN account_ch ON account_ch.uid = t.uid
    """).createOrReplaceTempView('pay_profiles')

    spark.sql(u"""
    SELECT
        type_name AS type_name,
        userid AS uid,
        CAST(utc8 as DATE) AS date,
        sid,
        gid,
        channel AS ch
    FROM jsonTable
    """).cache().createOrReplaceTempView('allevents')


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

# 注册7天内登陆不低于3天的用户中在第8-14天内登陆1天以上的用户数/
# 注册7天内登录不低于3天的用户数（用户样本为首周新注册）
def get_tencent2(spark):
    df = spark.sql(u'''
    WITH f AS (
        SELECT
           creates.uid AS uid,
           first(creates.date) AS date,
           COUNT(DISTINCT future_activity.date) AS times
        FROM creates_profiles AS creates
        LEFT JOIN allevents AS future_activity
        ON creates.uid = future_activity.uid
            AND DATEDIFF(future_activity.date, creates.date) <=7
        GROUP BY creates.uid, creates.date
    ), a3 AS (
        SELECT
            date, --
            COUNT(DISTINCT uid) AS a3
        FROM f
        WHERE times >= 3
        GROUP BY date --
    ), a8uid AS (
        SELECT
           creates.uid AS uid,
           first(creates.date) AS date,
           COUNT(DISTINCT future_activity.date) AS times
        FROM f AS creates
        LEFT JOIN allevents AS future_activity
        ON creates.uid = future_activity.uid
            AND DATEDIFF(future_activity.date, creates.date) >7
            AND DATEDIFF(future_activity.date, creates.date) <=14
        GROUP BY creates.uid, creates.date
    ), a14 AS (
        SELECT
            date,--
            COUNT(DISTINCT uid) AS a14
        FROM a8uid
        WHERE times > 0
        GROUP BY date --
    )
    --SELECT *, a14/a3 FROM a3 JOIN a14
    SELECT
        a3.date,
        a3.a3, a14.a14,
        (a14.a14/a3.a3) AS tencent2
    FROM a3
    LEFT JOIN a14
    ON a3.date = a14.date
    ''')
    return df.toPandas()

def get_tencent(spark):
    df = spark.sql(u'''
    WITH f AS (
        SELECT
           creates.uid AS uid,
           first(creates.date) AS date,
           COUNT(DISTINCT future_activity.date) AS times,
           SUM(DATEDIFF(future_activity.date,creates.date)) AS period
        FROM creates_profiles AS creates
        LEFT JOIN allevents AS future_activity
        ON creates.uid = future_activity.uid
            AND DATEDIFF(future_activity.date, creates.date) <=7
        GROUP BY creates.uid, creates.date
    ), a2 AS (
        SELECT
            date,
            COUNT(DISTINCT uid) AS a2
        FROM f
        WHERE times >= 2
        GROUP BY date
    ), a3 AS (
        SELECT
            date,
            COUNT(DISTINCT uid) AS a3
        FROM f
        WHERE times >= 3
        GROUP BY date
    )
    SELECT a2.date,
    a2.a2, a3.a3,
    (a3.a3/a2.a2) AS tencent
    FROM a2
    LEFT JOIN a3
    ON a2.date = a3.date
    ''')
    return df.toPandas()

def get_cohort_loginevent(spark):
    df = spark.sql(u'''
    WITH cohort_active_user_count AS (
        SELECT date, COUNT(DISTINCT uid) as count
        FROM creates_profiles
        GROUP BY date
    )
    SELECT date, period, -- 如何能够修改period为Day 01的形式？或者用topandas来实现
      new_users, retained_users, retention
    FROM (
        SELECT
           creates.date AS date,
           DATEDIFF(future_activity.date,creates.date) AS period,
           MAX(cohort_size.count) AS new_users,
           COUNT(distinct future_activity.uid) AS retained_users,
           COUNT(distinct future_activity.uid) / MAX(cohort_size.count) AS retention
        FROM creates_profiles AS creates
        LEFT JOIN loginsevents AS future_activity
        ON creates.uid = future_activity.uid
            AND creates.date <= future_activity.date --注意这里去掉=号，就没有day 0
            AND (creates.date + INTERVAL '14' DAY) >= future_activity.date
        LEFT JOIN cohort_active_user_count AS cohort_size
        ON creates.date = cohort_size.date
        GROUP BY creates.date, DATEDIFF(future_activity.date,creates.date)
    ) t
    WHERE period is not null
    ORDER BY date, period
    ''')
    return df.groupBy('date').pivot('period').max('retention').toPandas()


def get_cohort_allevent_spark(spark, profilesubset):
    #creates_profiles
    df = spark.sql(u'''
    WITH
        cohort_active_user_count AS (
        SELECT date, COUNT(DISTINCT uid) as count
        FROM {0}
        GROUP BY date
    )
    SELECT date, period,
      new_users, retained_users, retention
    FROM (
        SELECT
           creates.date AS date,
           DATEDIFF(future_activity.date,creates.date) AS period,
           MAX(cohort_size.count) AS new_users,
           COUNT(DISTINCT future_activity.uid) AS retained_users,
           COUNT(DISTINCT future_activity.uid) / MAX(cohort_size.count) AS retention
        FROM {0} AS creates
        LEFT JOIN allevents AS future_activity
        ON creates.uid = future_activity.uid
            AND creates.date < future_activity.date
            AND (creates.date + INTERVAL '14' DAY) >= future_activity.date
        LEFT JOIN cohort_active_user_count AS cohort_size
        ON creates.date = cohort_size.date
        GROUP BY creates.date, DATEDIFF(future_activity.date,creates.date)
    ) t
    WHERE period is not null
    ORDER BY date, period
    '''.format(profilesubset))
    # return df.groupBy('date').pivot('period').max('retention').toPandas()
    return df

# TODO by、sid?
def get_cohort_allevent_bych_spark(spark, profilesubset):
    df = spark.sql(u'''
    WITH
        cohort_active_user_count AS (
        SELECT date, ch, COUNT(DISTINCT uid) as count
        FROM {0}
        GROUP BY date, ch
    )
    SELECT date, period, ch,
      new_users, retained_users, retention
    FROM (
        SELECT
           creates.ch AS ch,
           creates.date AS date,
           DATEDIFF(future_activity.date,creates.date) AS period,
           MAX(cohort_size.count) AS new_users,
           COUNT(DISTINCT future_activity.uid) AS retained_users,
           COUNT(DISTINCT future_activity.uid) / MAX(cohort_size.count) AS retention
        FROM {0} AS creates
        LEFT JOIN allevents AS future_activity
        ON creates.uid = future_activity.uid
            AND creates.date < future_activity.date
            AND (creates.date + INTERVAL '14' DAY) >= future_activity.date
        LEFT JOIN cohort_active_user_count AS cohort_size
        ON creates.date = cohort_size.date AND creates.ch = cohort_size.ch
        GROUP BY creates.ch, creates.date, DATEDIFF(future_activity.date,creates.date)
    ) t
    WHERE period is not null
    ORDER BY date, period
    '''.format(profilesubset))
    # return df.groupBy('date').pivot('period').max('retention').toPandas()
    return df

'''
日新增&日活跃
'''
def get_daily_count(spark):
    df = spark.sql(u'''
    WITH daily_activities AS (
        SELECT date, COUNT(DISTINCT uid) AS daily_active
        FROM (
            SELECT userid AS uid, CAST(utc8 as DATE) AS date  FROM jsonTable
        ) t
        GROUP BY date
        ORDER BY date
    ), daily_new AS (
        SELECT date, COUNT(DISTINCT uid) AS daily_create
        FROM creates_profiles
        GROUP BY date
        ORDER BY date
    )
    SELECT daily_activities.date AS date, daily_create, daily_active
    FROM daily_activities
    LEFT JOIN daily_new
    ON daily_activities.date = daily_new.date
    ''')
    return df.toPandas()

'''
日增&日登 分渠道 结果
'''
def get_daily_count_bych(spark):
    df = spark.sql(u'''
    WITH daily_activities AS (
        SELECT date, ch ,COUNT(DISTINCT uid) AS daily_active
        FROM allevents
        GROUP BY date, ch
        ORDER BY date, ch
    ), daily_new AS (
        SELECT date, ch, COUNT(DISTINCT uid) AS daily_create
        FROM creates_profiles
        GROUP BY date, ch
        ORDER BY date, ch
    )
    SELECT
        daily_activities.date AS date,
        daily_activities.ch AS ch,
        daily_create,
        daily_active
    FROM daily_activities
    LEFT JOIN daily_new
    ON daily_activities.date = daily_new.date
        AND daily_activities.ch = daily_new.ch
    WHERE daily_activities.ch is not null
    ''')
    return df.toPandas()

'''
日增&日登 分服 结果
'''
def get_daily_count_bysid(spark):
    df = spark.sql(u'''
    WITH daily_activities AS (
        SELECT date, sid ,COUNT(DISTINCT uid) AS daily_active
        FROM allevents
        GROUP BY date, sid
        ORDER BY date, sid
    ), daily_new AS (
        SELECT date, sid, COUNT(DISTINCT uid) AS daily_create
        FROM creates_profiles
        GROUP BY date, sid
        ORDER BY date, sid
    )
    SELECT
        daily_activities.date AS date,
        daily_activities.sid AS sid,
        daily_create,
        daily_active
    FROM daily_activities
    LEFT JOIN daily_new
    ON daily_activities.date = daily_new.date
        AND daily_activities.sid = daily_new.sid
    ''')
    return df.toPandas()

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

    #tecent = get_tencent(spark)
    #upload_to_s3(tecent, 'cn-north-1', bucketname, '/%s/tencent.csv'%target_path)

    # tecent2 = get_tencent2(spark)
    # upload_to_s3(tecent2, 'cn-north-1', bucketname, '/%s/tencent2.csv'%target_path)
    # upload_to_s3(get_cohort_loginevent(spark), 'cn-north-1', bucketname, '/%s/cohort_login.csv'%target_path)

    cohort_xls_out = StringIO.StringIO()
    cohort_xls = pd.ExcelWriter(cohort_xls_out, engine='xlsxwriter')
    def cohort(targetprofiles, sheet_prefix):
        cohort_df = get_cohort_allevent_spark(spark, targetprofiles)
        cohort_df.groupBy('date').pivot('period').max('retention').toPandas().to_excel(cohort_xls, sheet_name=sheet_prefix+'retention_dayn')

        cohort_ch_df = get_cohort_allevent_bych_spark(spark, targetprofiles)
        cohort_ch_pandas = cohort_ch_df.toPandas()
        cohort_ch_pandas.loc[:,'ch']= cohort_ch_pandas.loc[:,'ch'].fillna(0).astype(float).astype(int)
        cohort_ch_pandas.to_excel(cohort_xls, sheet_name=sheet_prefix+'raw_ch_retention_dayn')

        #w = pd.merge(cohort_ch_pandas, channels, on='ch', how='left')
        w = cohort_ch_pandas
        crbych = pd.pivot_table(w, values=['new_users','retained_users','retention'],
                   index=['ch','date'], columns=['period'], aggfunc=np.sum)
        crbych2 = pd.pivot_table(w, values=['new_users','retained_users','retention'],
                   index=['date','ch'], columns=['period'], aggfunc=np.sum)
        try:
            cohort_result_bych = crbych.stack(0).unstack(2)
            cohort_result_bych.to_excel(cohort_xls, sheet_name=sheet_prefix+'retention_dayn_ch')

            cohort_result_bych2 = crbych2.stack(0).unstack(2)
            cohort_result_bych2.to_excel(cohort_xls, sheet_name=sheet_prefix+'retention_dayn_ch2')
        except:
            pass
    cohort('creates_profiles', '')
    cohort('pay_profiles', 'iap_')
    cohort_xls.save()
    upload_stringio_to_s3(cohort_xls_out, 'cn-north-1', bucketname, '/%s/retention_dayn.xlsx'%target_path)


    upload_to_s3(get_daily_count(spark), 'cn-north-1', bucketname, '/%s/daily_activities.csv'%target_path)
    upload_to_s3(get_daily_count_bych(spark), 'cn-north-1', bucketname, '/%s/daily_activities_by_ch.csv'%target_path)
    upload_to_s3(get_daily_count_bysid(spark), 'cn-north-1', bucketname, '/%s/daily_activities_by_sid.csv'%target_path)    

    # #level funnel
    # leveldf = get_level_funnel_by_ch(spark)
    # ndf = pd.pivot_table(leveldf, values=['events_accounts'],
    #            index=['event'], columns=['ch', 'date'], aggfunc=np.sum)
    # ndf2 = pd.pivot_table(leveldf, values=['events_accounts'],
    #            index=['event'], columns=['date'], aggfunc=np.sum)

    # levelfunnel_xls_out = StringIO.StringIO()
    # levelfunnel_xls = pd.ExcelWriter(levelfunnel_xls_out, engine='xlsxwriter')
    # ndf.to_excel(levelfunnel_xls, sheet_name='level_ch')
    # ndf2.to_excel(levelfunnel_xls, sheet_name="funnel_create")
    # get_level_funnel(spark).to_excel(levelfunnel_xls, sheet_name="funnel_all")
    # levelfunnel_xls.save()
    # upload_stringio_to_s3(levelfunnel_xls_out, 'cn-north-1', bucketname, '/%s/level_funnel.xlsx'%target_path)

    spark.stop()

    #livedata@taiyouxi.cn
    sys.exit(email_report('livedata@taiyouxi.cn', u'IFSG 数据留存计算 %s'%(' '.join(sys.argv[3:])), u'''
    IFSG 数据留存计算
    测试数据效果
    %s
    '''%(' '.join(sys.argv))))
