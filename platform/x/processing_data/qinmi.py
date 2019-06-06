# -*- coding: utf-8 -*-
import json
import os
import csv
import xlwt
import datetime
import sys
from boto import s3
import MySQLdb  # 载入连接数据库模块

title = ["Date","Gid","ServerID","NewAccountNum","GunAccountNum","ActiveNum","ActiveOldNum","Pay",
         "PayNum","ActivePay","PayARPU","PayARPPU","ActiveARPU","NewPay","NewPayNum",
         "NewGUn","NewGUnPay","NewAccountPay","NewAccountARPU"]

#[0.[大于30级的uid],1.[当天登入的uid],2.[当天新注册的UID],3.[滚服账号uid],4.[付费金额],5.[付费账号uid],
# 6.[新增账号付费人数],7.[新增账号付费金额],8.[滚服新增付费人数],9.[滚服新增付费金额],10[大区]]
def infomation(userInfoDataPath,chargeDataPath,strTime):
    print "begin"
    wordOne = "userinfo"
    dataUser = search(userInfoDataPath, wordOne)
    wordTwo = "charge"
    datacharge = search(chargeDataPath, wordTwo)
    informationData = {}
    thirty = []
    serverTime = {}
    today = ""
    try:  # 尝试连接数据库
        conn = MySQLdb.connect(host="54.223.192.252", user="test1", passwd="QmPhaQ8hYsxx", db="test")
    except MySQLdb.OperationalError, message:  # 连接失败提示
        print "link error"
    cursor = conn.cursor()  # 定义连接对象
    cursor.execute("select sid,start_time from server_startTime")  # 使用cursor提供的方法来执行查询语句
    data = cursor.fetchall()  # 使用fetchall方法返回所有查询结果
    for x in data:
        serverTime.setdefault(x[0],x[1])

    for file in dataUser:
        xxxx = file.split('/')
        ffff = xxxx[len(xxxx) - 1].split('_')
        today = ffff[len(ffff) - 1][0:10]
        print ffff
        with open(file) as f:
            reader = csv.reader(f)
            for line in reader:
                if line[0] == "accountId":
                    vip = line.index("VIP等级")
                    cLvl = line.index("战队等级")
                    gs = line.index("战队总战力")
                    hc = line.index("硬通数")
                    new= line.index("注册时间")
                    lastLogginIn = line.index("最后登陆日")
                    continue
                try:
                    temp = line[0].split(':')
                    if temp < 3:
                        print temp
                        continue
                    gid = temp[0]
                    uid = temp[2]
                    sid = temp[1]
                    if sid not in informationData.keys():
                        informationData.setdefault(sid,[[],[],[],[],[],[],[],[],[],[],[],[]])
                        informationData[sid][10].append(gid)
                        serverStartTime = getRelativedays(serverTime[sid],strTime)
                        informationData[sid][11].append(serverStartTime)
                    if int(line[cLvl]) >= 30:
                        thirty.append(uid)
                    if line[lastLogginIn][0:10] == ffff[len(ffff) - 1][0:10]:
                        informationData[sid][1].append(uid)
                    if line[new][0:10] == ffff[len(ffff) - 1][0:10]:
                        informationData[sid][2].append(uid)
                except Exception,e:
                    print temp


    print "30Clv people",len(thirty)
    for key in informationData:
        print "newnew", key, informationData[key][2]
        for subkey in range(len(informationData[key][2])):
            if informationData[key][2][subkey] in thirty:
                print "log",informationData[key][2][subkey],key
                informationData[key][3].append(informationData[key][2][subkey])

    print "read charge"
    for file in datacharge:
        print file
        with open(file) as f:
            for line in f:
                if line.startswith('$$'):
                    print "#####", x,file
                    continue
                x = line.split('$$')
                if len(x) < 2:
                    print "line < 2",x
                    continue
                temp = x[4].split(":")

                if len(temp) <2:
                    continue
                sid = temp[1]
                uid = temp[2]
                pay = 0
                if x[8].isdigit():
                    pay = (int(x[8]))
                else:
                    pay = (float(x[8]))
                informationData[sid][4].append(pay)
                if uid not in informationData[sid][5]:
                    informationData[sid][5].append(uid)
                if uid in informationData[sid][2]:
                    informationData[sid][6].append(uid)
                    informationData[sid][7].append(pay)
                if uid in informationData[sid][3]:
                    informationData[sid][8].append(uid)
                    informationData[sid][9].append(pay)




    sortltvData = [(k, informationData[k]) for k in sorted(informationData.keys())]
    # 写入数据
    print "write now"
    workbook = xlwt.Workbook(encoding='ascii')
    worksheet = workbook.add_sheet('qin')

    for t in range(len(title)):
        worksheet.write(0,t,title[t])
    count = 1
    for key in sortltvData:
        # [0.[大于30级的uid],1.[当天登入的uid],2.[当天新注册的UID],3.[滚服账号uid],4.[付费金额],5.[付费账号uid],
        # 6.[新增账号付费人数],7.[新增账号付费金额],8.[滚服新增付费人数],9.[滚服新增付费金额]]

        worksheet.write(count, 0, today)       #日期
        worksheet.write(count, 1, key[1][10])  #大区
        worksheet.write(count, 2, key[0])      #区号
        worksheet.write(count, 3, len(key[1][2])) #新注册
        worksheet.write(count, 4, len(key[1][3]))      #滚服新账号数
        worksheet.write(count, 5, len(key[1][1]))      #活跃账号数
        worksheet.write(count, 6, len(key[1][1]) - len(key[1][2]))      #活跃老账号数
        worksheet.write(count, 7, sum(key[1][4]))      #付费金额
        worksheet.write(count, 8, len(key[1][5]))      #付费账号数

        if len(key[1][5]) <= 0 or len(key[1][1]) <= 0:
            worksheet.write(count, 9, 0)
        else:
            worksheet.write(count, 9, float(len(key[1][5])) / float(len(key[1][1])))      #活跃付费率

        if sum(key[1][4]) <= 0 or len(key[1][2]) <= 0:
            worksheet.write(count, 10, 0)
        else:
            worksheet.write(count, 10, float(sum(key[1][4])) / float(len(key[1][2])))      #ARPU

        if sum(key[1][4]) <= 0 or len(key[1][5]) <= 0:
            worksheet.write(count, 11, 0)
        else:
            worksheet.write(count, 11, float(sum(key[1][4])) / float(len(key[1][5])))      #ARPPU

        if sum(key[1][4]) <= 0 or len(key[1][1]) <= 0:
            worksheet.write(count, 12, 0)
        else:
            worksheet.write(count, 12, float(sum(key[1][4])) / float(len(key[1][1])))      #活跃ARPU

        worksheet.write(count, 13, len(key[1][6]))      #新增账号付费人数
        worksheet.write(count, 14, sum(key[1][7]))      #新增账号付费金额
        worksheet.write(count, 15, len(key[1][8]))      #滚服新增付费人数
        worksheet.write(count, 16, sum(key[1][9]))      #滚服新增付费金额

        if len(key[1][6]) <= 0 or len(key[1][2]) <=0:
            worksheet.write(count, 17, 0)
        else:
            worksheet.write(count, 17, float(len(key[1][6])) / float(len(key[1][2])))      #新增账号付费率
        if sum(key[1][7]) <= 0 or len(key[1][2]) <= 0:
            worksheet.write(count, 18, 0)
        else:
            worksheet.write(count, 18, float(sum(key[1][7])) / float(len(key[1][2])))      #新增账号付费ARPU
        if len(key[1][11]) == 0:
            worksheet.write(count, 19, "-1")
            print key[0]
        else:
            worksheet.write(count, 19,str(key[1][11][0]))
        count += 1
    workbook.save("./data.xls")
    print "end",datetime.datetime.now()

def search(path, word):
    name = []
    for filename in os.listdir(path):
        fp = os.path.join(path, filename)
        if os.path.isfile(fp) and word in filename:
            name.append(fp)
        elif os.path.isdir(fp):
            search(fp, word)
    return name


def store(jName,data):
    with open(jName, 'w') as json_file:
        json_file.write(data)

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
    d1 = datetime.date(int(s[0]), int(s[1]), int(s[2]))
    d2 = datetime.date(int(e[0]), int(e[1]), int(e[2]))
    return ((d2 - d1).days)


def as_attachments(name, content):
    global attachments
    attachments.append(('attachment', (name, content)))
def get_user_event_info(data, daterange):

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
            if a == data and str(logsp[2]) == 'userinfo' and logsp[len(logsp)-1].split('.')[1] == 'csv':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retUser.append(k.name)
            if a == data and str(logsp[2]) == 'charge' and logsp[len(logsp)-1].split('.')[1] == 'log':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retPay.append(k.name)
    print('total:%d'%(total_size/1024.0/1024.0/1024.0))
    # print retUser
    # print retPay
    print "im here1"
    if not os.path.exists(daterange+'/bilogs-csv/'):
        print "im here2"
        os.makedirs(daterange+'/bilogs-csv/')
    if not os.path.exists(daterange+'/charge/'):
        print "im here3"
        os.makedirs(daterange+'/charge/')
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
            print (dPath + '/bilogs-csv/' + name)
            key.get_contents_to_filename(dPath + '/bilogs-csv/' + name)
        else:
            print (dPath + '/charge/' + name)
            key.get_contents_to_filename(dPath + '/charge/' + name)

if __name__ == '__main__':
    today = datetime.date.today()
    yesterday = today - datetime.timedelta(days=1)
    strYesterday = yesterday.strftime('%Y-%m-%d')
    #get_user_event_info(strYesterday, "./s3-data/"+strYesterday)
    infomation("./s3-data/"+strYesterday+"/bilogs-csv/","./s3-data/"+strYesterday+"/charge/",strYesterday)
    with open("./data.xls") as f:
        as_attachments(os.path.basename("./data.xls"), f)
        # livedata@taiyouxi.cn
        sys.exit(email_report('huquanqi@taiyouxi.cn', u'分服新增、付费、arpu存量数据表 %s' % (strYesterday), u'''
       分服新增、付费、arpu存量数据表
       %s
       ''' % (strYesterday)))