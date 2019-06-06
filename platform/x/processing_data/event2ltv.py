# -*- coding: utf-8 -*-

import json
import os
import datetime
import xlwt
import sys
import csv
from boto import s3
import MySQLdb

def updateUid2CreatTime(userInfoDataPath):

    dataUser = search(userInfoDataPath)

    # data200 = {}
    # data201 = {}
    # data202 = {}
    # data203 = {}
    with open("./uid2Time_200.json") as f:
        data200 = json.load(f)
    with open("./uid2Time_201.json") as f:
        data201 = json.load(f)
    with open("./uid2Time_202.json") as f:
        data202 = json.load(f)
    with open("./uid2Time_203.json") as f:
        data203 = json.load(f)

    fileNum = 0
    for file in dataUser:
        fileNum += 1
        print "###", file, fileNum
        with open(file) as f:
            reader = csv.reader(f)
            for line in reader:
                if len(line) <=0:
                    continue
                if line[0] == "accountId":
                    rTime = line.index("注册时间")
                    continue
                try:
                    tmp = line[0].split(':')
                    uid = tmp[2]
                    if line[0][0:3] == '200':
                        if data200.has_key(uid):
                            if line[rTime][0:10] < data200[uid]:
                                data200[uid] = line[rTime][0:10]
                        else:
                            data200.setdefault(uid, line[rTime][0:10])

                    if line[0][0:3] == '201':
                        if data201.has_key(uid):
                            if line[rTime][0:10] < data201[uid]:
                                data201[uid] = line[rTime][0:10]
                        else:
                            data201.setdefault(uid, line[rTime][0:10])

                    if line[0][0:3] == '202':
                        if data202.has_key(uid):
                            if line[rTime][0:10] < data202[uid]:
                                data202[uid] = line[rTime][0:10]
                        else:
                            data202.setdefault(uid, line[rTime][0:10])

                    if line[0][0:3] == '203':
                        if data203.has_key(uid):
                            if line[rTime][0:10] < data203[uid]:
                                data203[uid] = line[rTime][0:10]
                        else:
                            data203.setdefault(uid, line[rTime][0:10])
                except Exception, e:
                    print Exception, e
                    print line[0], file

    json_str200 = json.dumps(data200)
    json_str201 = json.dumps(data201)
    json_str202 = json.dumps(data202)
    json_str203 = json.dumps(data203)
    store("uid2Time_200.json", json_str200)
    store("uid2Time_201.json", json_str201)
    store("uid2Time_202.json", json_str202)
    store("uid2Time_203.json", json_str203)

def acid2CreatTime(chargeDataPath):
    wordTwo = "charge"
    datacharge = search(chargeDataPath)
    print datacharge
    # ltvData200 = {}
    # ltvData201 = {}
    # ltvData202 = {}
    # ltvData203 = {}
    with open("./uid2Time_200.json") as f:
        data200 = json.load(f)
    with open("./uid2Time_201.json") as f:
        data201 = json.load(f)
    with open("./uid2Time_202.json") as f:
        data202 = json.load(f)
    with open("./uid2Time_203.json") as f:
        data203 = json.load(f)

    with open("./ltvData_200.json") as f:
        ltvData200 = json.load(f)
    with open("./ltvData_201.json") as f:
        ltvData201 = json.load(f)
    with open("./ltvData_202.json") as f:
        ltvData202 = json.load(f)
    with open("./ltvData_203.json") as f:
        ltvData203 = json.load(f)

    print "### READ ok"
    for file in datacharge:
        with open(file) as f:
            for line in f:
                if line.startswith('$$'):
                    print "#####", x,file
                    continue
                x = line.split('$$')
                if len(x) < 2:
                    print "line < 2",x
                    continue
                sid = x[1].split(":")
                print sid
                if len(sid) < 2:
                    continue
                uid = sid[1]
                try:
                    if len(x) < 5:
                        print "#####", x,file
                        continue


                    if sid[0] == "200":
                        if data200.has_key(uid):
                            creatTime = data200[uid]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData200.keys():
                            ltvData200.setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData200[creatTime].keys():
                            ltvData200[creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData200[creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData200[creatTime][x[11][0:10]].append(float(x[8]))
                    if sid[0] == "201":
                        if data201.has_key(uid):
                            creatTime = data201[uid]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData201.keys():
                            ltvData201.setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData201[creatTime].keys():
                            ltvData201[creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData201[creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData201[creatTime][x[11][0:10]].append(float(x[8]))

                    if sid[0] == "202":
                        if data202.has_key(uid):
                            creatTime = data202[uid]
                        else:
                            continue
                        if len(creatTime) > 15:
                            continue
                        if creatTime not in ltvData202.keys():
                            ltvData202.setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData202[creatTime].keys():
                            ltvData202[creatTime].setdefault(x[11][0:10],[])

                        if x[8].isdigit():
                            ltvData202[creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData202[creatTime][x[11][0:10]].append(float(x[8]))

                    if sid[0] == "203":
                        if data203.has_key(uid):
                            creatTime = data203[uid]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData203.keys():
                            ltvData203.setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData203[creatTime].keys():
                            ltvData203[creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData203[creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData203[creatTime][x[11][0:10]].append(int(float(x[8])))
                except Exception,e:
                    print Exception,e
                    print "linlin",line
    try:
        for key in ltvData200:
            for subkey in ltvData200[key]:
                print "###",key,subkey,sum(ltvData200[key][subkey])
    except Exception,e:
        print Exception,e
        print "wrong",ltvData200[key][subkey]

    json_str200 = json.dumps(ltvData200)
    json_str201 = json.dumps(ltvData201)
    json_str202 = json.dumps(ltvData202)
    json_str203 = json.dumps(ltvData203)
    store("ltvData_200.json",json_str200)
    store("ltvData_201.json",json_str201)
    store("ltvData_202.json",json_str202)
    store("ltvData_203.json",json_str203)

    sortltvData200 = [(k,ltvData200[k]) for k in sorted(ltvData200.keys())]
    sortltvData201 = [(k,ltvData201[k]) for k in sorted(ltvData201.keys())]
    sortltvData202 = [(k,ltvData202[k]) for k in sorted(ltvData202.keys())]
    sortltvData203 = [(k,ltvData203[k]) for k in sorted(ltvData203.keys())]


    # sortltvData200.sort()
    # sortltvData201.sort()
    # sortltvData202.sort()
    # sortltvData203.sort()
    title = []
    for key in sortltvData200:
        if key[0] < "2016-11-10":
            continue
        title.append(key[0])

    database = MySQLdb.connect(host="54.223.192.252", user="test1", passwd="QmPhaQ8hYsxx", db="test")
    cursor = database.cursor()

    # 创建插入SQL语句
    queryByGid = """INSERT INTO ltv_byGid (gid,creat_time,people_count,days,consume)
     VALUES (%s, %s, %s, %s, %s)"""



    #写入数据
    print "write now"
    workbook = xlwt.Workbook(encoding='ascii')
    worksheet200 = workbook.add_sheet('ltv200')
    worksheet201 = workbook.add_sheet('ltv201')
    worksheet202 = workbook.add_sheet('ltv202')
    worksheet203 = workbook.add_sheet('ltv203')


    for t in range(len(title)):
        worksheet200.write(0,t+2,t+1)
        worksheet201.write(0,t+2,t+1)
        worksheet202.write(0,t+2,t+1)
        worksheet203.write(0,t+2,t+1)
    count = 0
    subcount = 1
    peoples = 0
    paySum = 0
    try:
        for key in sortltvData200:
            if key[0] < "2016-11-10":
                continue
            if len(key[0]) > 10:
                continue
            count +=1
            worksheet200.write(count, 0, key[0])

            for acidkey in data200:
                if data200[acidkey] == key[0]:
                    peoples += 1

            worksheet200.write(count, 1, peoples)

            sortkey = [(k,key[1][k]) for k in sorted(key[1].keys())]

            for subkey in sortkey:
                if subkey[0] < "2016-11-10":
                    continue
                if len(subkey[0]) > 10:
                    continue
                subcount += 1
                paySum += sum(subkey[1])

                day = getRelativedays(key[0],subkey[0])

                #worksheet200.write(count,subcount,sum(subkey[1]))
                worksheet200.write(count, day + 2, float(paySum) / float(peoples))
                values = ("200",str(key[0]),str(peoples),str(day+1),str(float(paySum) / float(peoples)))
                # 执行sql语句
                cursor.execute(queryByGid, values)
            paySum = 0
            subcount = 1
            peoples = 0
    except Exception, e:
        print "200 write wrong",Exception, e
        print key[0],"####",key[1]
        print "201day", day, key[0], subkey[0],subkey

    count1 = 0
    subcount1 = 1
    peoples1 = 0
    paySum1 = 0
    try:
        for key in sortltvData201:
            if key[0] < "2016-11-10":
                continue
            if len(key[0]) > 10:
                continue
            count1 +=1
            worksheet201.write(count1, 0, key[0])

            for acidkey in data201:
                if data201[acidkey] == key[0]:
                    peoples1 += 1

            worksheet201.write(count1, 1, peoples1)
            sortkey = [(k, key[1][k]) for k in sorted(key[1].keys())]

            for subkey in sortkey:
                if subkey[0] < "2016-11-10":
                    continue
                if len(subkey[0]) > 10:
                    continue
                if subkey[0] < key[0]:
                    continue
                subcount1 += 1
                paySum1 += sum(subkey[1])

                day = getRelativedays(key[0], subkey[0])

                #worksheet201.write(count1, subcount1, sum(subkey[1]))
                worksheet201.write(count1, day + 2, float(paySum1) / float(peoples1))
                values = ("201",str(key[0]),str(peoples1),str(day+1),str(float(paySum1) / float(peoples1)))
                # 执行sql语句
                cursor.execute(queryByGid, values)
            paySum1 = 0
            subcount1 = 1
            peoples1 = 0
    except Exception,e:
        print "201day",day,key[0],subkey[0],subkey

    count2 = 0
    subcount2 = 1
    peoples2 = 0
    paySum2 = 0
    try:
        for key in sortltvData202:
            if key[0] < "2016-11-10":
                continue
            if len(key[0]) > 10:
                continue
            if subkey[0] < key[0]:
                continue
            count2 +=1

            worksheet202.write(count2, 0, key[0])

            for acidkey in data202:
                if data202[acidkey] == key[0]:
                    peoples2 += 1

            worksheet202.write(count2, 1, peoples2)

            sortkey = [(k, key[1][k]) for k in sorted(key[1].keys())]

            for subkey in sortkey:
                if subkey[0] < "2016-11-10":
                    continue
                if len(subkey[0]) > 10:
                    continue
                subcount2 += 1
                paySum2 += sum(subkey[1])

                day = getRelativedays(key[0],subkey[0])

                #worksheet202.write(count2, subcount2, sum(subkey[1]))
                worksheet202.write(count2, day + 2, float(paySum2) / float(peoples2))
                values = ("202",str(key[0]),str(peoples2),str(day+1),str(float(paySum2) / float(peoples2)))
                # 执行sql语句
                cursor.execute(queryByGid, values)
            paySum2 = 0
            subcount2 = 1
            peoples2 = 0
    except Exception,e:
        print "202day",day,key[0],subkey[0],subkey

    count3 = 0
    subcount3 = 1
    peoples3 = 0
    paySum3 = 0
    try:
        for key in sortltvData203:
            if key[0] < "2016-11-10":
                continue
            if len(key[0]) > 10:
                continue
            if subkey[0] < key[0]:
                continue
            count3 +=1
            worksheet203.write(count3, 0, key[0])

            for acidkey in data203:
                if data203[acidkey] == key[0]:
                    peoples3 += 1

            worksheet203.write(count3, 1, peoples3)
            sortkey = [(k, key[1][k]) for k in sorted(key[1].keys())]

            for subkey in sortkey:
                if subkey[0] < "2016-11-10":
                    continue
                if len(subkey[0]) > 10:
                    continue
                subcount3 += 1
                paySum3 += sum(subkey[1])

                day = getRelativedays(key[0],subkey[0])

                worksheet203.write(count3, day+2, float(paySum3) / float(peoples3))
                values = ("203",str(key[0]),str(peoples3),str(day+1),str(float(paySum3) / float(peoples3)))
                # 执行sql语句
                cursor.execute(queryByGid, values)
                #worksheet203.write(count3, subcount3, sum(subkey[1]))
            paySum3 = 0
            subcount3 = 1
            peoples3 = 0
    except Exception,e:
        print "203day",day,key[0],subkey[0],subkey

    # 关闭游标
    cursor.close()

    # 提交
    database.commit()

    # 关闭数据库连接
    database.close()

    workbook.save("/home/ec2-user/lty.xls")
    print "end",datetime.datetime.now()


def search(path):
    name = []
    for filename in os.listdir(path):
        fp = os.path.join(path, filename)
        if os.path.isfile(fp):
            name.append(fp)
        elif os.path.isdir(fp):
            search(fp)
    return name


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
    # today = datetime.date.today()
    # yesterday = today - datetime.timedelta(days=1)
    # strYesterday = yesterday.strftime('%Y-%m-%d')
    strYesterday = sys.argv[1]
    get_user_event_info(strYesterday,"./s3-ltv-linshi/"+strYesterday)
    updateUid2CreatTime("./s3-ltv-linshi/"+strYesterday+"/bilogs-csv/")
    print "Uid2CreatTime update ok"
    acid2CreatTime("./s3-ltv-linshi/"+strYesterday+"/charge/")

    with open("/home/ec2-user/lty.xls") as f:
        as_attachments(os.path.basename("/home/ec2-user/lty.xls"), f)
        # livedata@taiyouxi.cn
        sys.exit(email_report('livedata@taiyouxi.cn', u'ltv 数据 %s' % (strYesterday), u'''
       ltv 数据
       %s
       ''' % (strYesterday)))