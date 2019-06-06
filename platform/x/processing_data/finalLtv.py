# -*- coding: utf-8 -*-

import json
import os
import datetime
import xlwt
from boto import s3
import csv
import sys


def acid2CreatTime(userInfoDataPath,chargeDataPath):
    wordOne = "userinfo"
    wordTwo = "charge"
    dataUser = search(userInfoDataPath, wordOne)
    datacharge = search(chargeDataPath,wordTwo)
    # ltvData_200 = {}
    # ltvData_201 = {}
    # ltvData_202 = {}
    # ltvData_203 = {}
    with open("/home/ec2-user/acid2Time_200.json") as f:
        data200 = json.load(f)
    with open("/home/ec2-user/acid2Time_201.json") as f:
        data201 = json.load(f)
    with open("/home/ec2-user/acid2Time_202.json") as f:
        data202 = json.load(f)
    with open("/home/ec2-user/acid2Time_203.json") as f:
        data203 = json.load(f)

    with open("/home/ec2-user/ltvData_200.json") as f:
        ltvData_200 = json.load(f)
    with open("/home/ec2-user/ltvData_201.json") as f:
        ltvData_201 = json.load(f)
    with open("/home/ec2-user/ltvData_202.json") as f:
        ltvData_202 = json.load(f)
    with open("/home/ec2-user/ltvData_203.json") as f:
        ltvData_203 = json.load(f)
    print "### READ ok"
    fileNum = 0
    for file in dataUser:
        fileNum += 1
        print file,fileNum,datetime.datetime.now()
        with open(file) as f:
            reader = csv.reader(f)
            for line in reader:
                if line[0] == "accountId":
                    rTime = line.index("注册时间")
                    continue
                if line[0][0:3] == '200':
                    if data200.has_key(line[0]):
                        if line[rTime][0:10] < data200[line[0]]:
                            data200[line[0]] = line[rTime]
                    else:
                        data200.setdefault(line[0],line[rTime][0:10])

                if line[0][0:3] == '201':
                    if data201.has_key(line[0]):
                        if line[rTime][0:10] < data201[line[0]]:
                            data201[line[0]] = line[rTime]
                    else:
                        data201.setdefault(line[0],line[rTime][0:10])

                if line[0][0:3] == '202':
                    if data202.has_key(line[0]):
                        if line[rTime][0:10] < data202[line[0]]:
                            data202[line[0]] = line[rTime]
                    else:
                        data202.setdefault(line[0],line[rTime][0:10])

                if line[0][0:3] == '203':
                    if data203.has_key(line[0]):
                        if line[rTime][0:10] < data203[line[0]]:
                            data203[line[0]] = line[rTime]
                    else:
                        data203.setdefault(line[0],line[rTime][0:10])


    json_str200 = json.dumps(data200)
    json_str201 = json.dumps(data201)
    json_str202 = json.dumps(data202)
    json_str203 = json.dumps(data203)
    store("acid2Time_200.json",json_str200)
    store("acid2Time_201.json",json_str201)
    store("acid2Time_202.json",json_str202)
    store("acid2Time_203.json",json_str203)


    for file in datacharge:
        with open(file) as f:
            for line in f:
                if line.startswith('$$'):
                    print "#####", x,file
                    continue
                x = line.split('$$')
                sid = x[1].split(":")
                try:
                    if len(x) < 5:
                        print "#####", x,file
                        continue


                    if sid[0] == "200":
                        if data200.has_key(x[4]):
                            creatTime = data200[x[4]]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData_200.keys():
                            ltvData_200.setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData_200[creatTime].keys():
                            ltvData_200[creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData_200[creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData_200[creatTime][x[11][0:10]].append(float(x[8]))
                    if sid[0] == "201":
                        if data201.has_key(x[4]):
                            creatTime = data201[x[4]]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData_201.keys():
                            ltvData_201.setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData_201[creatTime].keys():
                            ltvData_201[creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData_201[creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData_201[creatTime][x[11][0:10]].append(float(x[8]))

                    if sid[0] == "202":
                        if data202.has_key(x[4]):
                            creatTime = data202[x[4]]
                        else:
                            continue
                        if len(creatTime) > 15:
                            continue
                        if creatTime not in ltvData_202.keys():
                            ltvData_202.setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData_202[creatTime].keys():
                            ltvData_202[creatTime].setdefault(x[11][0:10],[])

                        if x[8].isdigit():
                            ltvData_202[creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData_202[creatTime][x[11][0:10]].append(float(x[8]))

                    if sid[0] == "203":
                        if data203.has_key(x[4]):
                            creatTime = data203[x[4]]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData_203.keys():
                            ltvData_203.setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData_203[creatTime].keys():
                            ltvData_203[creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData_203[creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData_203[creatTime][x[11][0:10]].append(int(float(x[8])))
                except Exception,e:
                    print Exception,e
                    print "linlin",line
    try:
        for key in ltvData_200:
            for subkey in ltvData_200[key]:
                print "###",key,subkey,sum(ltvData_200[key][subkey])
    except Exception,e:
        print Exception,e
        print "wrong",ltvData_200[key][subkey]

    json_str200 = json.dumps(ltvData_200)
    json_str201 = json.dumps(ltvData_201)
    json_str202 = json.dumps(ltvData_202)
    json_str203 = json.dumps(ltvData_203)
    store("ltvData_200.json",json_str200)
    store("ltvData_201.json",json_str201)
    store("ltvData_202.json",json_str202)
    store("ltvData_203.json",json_str203)

    sortltvData200 = [(k,ltvData_200[k]) for k in sorted(ltvData_200.keys())]
    sortltvData201 = [(k,ltvData_201[k]) for k in sorted(ltvData_201.keys())]
    sortltvData202 = [(k,ltvData_202[k]) for k in sorted(ltvData_202.keys())]
    sortltvData203 = [(k,ltvData_203[k]) for k in sorted(ltvData_203.keys())]


    # sortltvData200.sort()
    # sortltvData201.sort()
    # sortltvData202.sort()
    # sortltvData203.sort()
    title = []
    for key in sortltvData200:
        title.append(key[0])


    #写入数据
    print "write now"
    workbook = xlwt.Workbook(encoding='ascii')
    worksheet200 = workbook.add_sheet('ltv200')
    worksheet201 = workbook.add_sheet('ltv201')
    worksheet202 = workbook.add_sheet('ltv202')
    worksheet203 = workbook.add_sheet('ltv203')


    for t in range(len(title)):
        worksheet200.write(0,t+1,title[t])
        worksheet201.write(0,t+1,title[t])
        worksheet202.write(0,t+1,title[t])
        worksheet203.write(0,t+1,title[t])
    count = 0
    subcount = 1
    peoples = 0
    try:
        for key in sortltvData200:
            print "200", key[0]
            count +=1
            worksheet200.write(count, 0, key[0])

            for acidkey in data200:
                if data200[acidkey] == key[0]:
                    peoples += 1

            worksheet200.write(count, 1, peoples)

            sortkey = [(k,key[1][k]) for k in sorted(key[1].keys())]
            for subkey in sortkey:
                subcount += 1
                worksheet200.write(count,subcount,sum(subkey[1]))
            subcount = 1
            peoples = 0
    except Exception, e:
        print "200 write wrong",Exception, e
        print key[0],"####",key[1]

    count1 = 0
    subcount1 = 1
    peoples1 = 0
    for key in sortltvData201:
        print "201",key[0]
        count1 +=1
        worksheet201.write(count1, 0, key[0])

        for acidkey in data201:
            if data201[acidkey] == key[0]:
                peoples1 += 1

        worksheet201.write(count1, 1, peoples1)
        sortkey = [(k, key[1][k]) for k in sorted(key[1].keys())]
        for subkey in sortkey:
            subcount1 += 1
            worksheet201.write(count1, subcount1, sum(subkey[1]))
        print peoples
        subcount1 = 1
        peoples1 = 0


    count2 = 0
    subcount2 = 1
    peoples2 = 0
    for key in sortltvData202:
        print "202",key[0]
        count2 +=1
        worksheet202.write(count2, 0, key[0])

        for acidkey in data202:
            if data202[acidkey] == key[0]:
                peoples2 += 1

        worksheet202.write(count2, 1, peoples2)

        sortkey = [(k, key[1][k]) for k in sorted(key[1].keys())]
        for subkey in sortkey:
            subcount2 += 1
            worksheet202.write(count2, subcount2, sum(subkey[1]))
        print peoples2
        subcount2 = 1
        peoples2 = 0

    count3 = 0
    subcount3 = 1
    peoples3 = 0
    for key in sortltvData203:
        print "203",key[0]
        count3 +=1
        worksheet203.write(count3, 0, key[0])

        for acidkey in data203:
            if data203[acidkey] == key[0]:
                peoples3 += 1

        worksheet203.write(count3, 1, peoples3)
        sortkey = [(k, key[1][k]) for k in sorted(key[1].keys())]
        for subkey in sortkey:
            subcount3 += 1
            # worksheet203.write(count3, title.index(subkey[0])+1, sum(subkey[1]))
            worksheet203.write(count3, subcount3, sum(subkey[1]))
        print peoples3
        subcount3 = 1
        peoples3 = 0

    workbook.save("/home/ec2-user/lty.xls")
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
if __name__ == '__main__':
    today = datetime.date.today()
    yesterday = today - datetime.timedelta(days=1)
    strYesterday = yesterday.strftime('%Y-%m-%d')

    #get_user_event_info(strYesterday,"/home/ec2-user/s3-ltv-linshi/"+strYesterday)
    acid2CreatTime("/home/ec2-user/s3-ltv-linshi/"+strYesterday+"/bilogs-csv","/home/ec2-user/s3-ltv-linshi/"+strYesterday+"/charge")
    with open("/home/ec2-user/ltyTest.xls") as f:
        as_attachments(os.path.basename("/home/ec2-user/ltyTest.xls"), f)
        # livedata@taiyouxi.cn
        sys.exit(email_report('huquanqi@taiyouxi.cn', u'VIP 数据留存计算 %s' % (strYesterday), u'''
       VIP 数据留存计算
       测试数据效果
       %s
       ''' % (strYesterday)))
