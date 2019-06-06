# -*- coding: utf-8 -*-
import json
import os
import csv
import xlwt
import datetime
import sys
from boto import s3

title = ["Date","Gid","ServerID","NewAccountNum","GunAccountNum","ActiveNum","ActiveOldNum","Pay",
         "PayNum","ActivePay","PayARPU","PayARPPU","ActiveARPU","NewPay","NewPayNum",
         "NewGUn","NewGUnPay","NewAccountPay","NewAccountARPU"]

#[0.[大于30级的uid],1.[当天登入的uid],2.[当天新注册的UID],3.[滚服账号uid],4.[付费金额],5.[付费账号uid],
# 6.[新增账号付费人数],7.[新增账号付费金额],8.[滚服新增付费人数],9.[滚服新增付费金额],10[大区]]
def infomation(userInfoDataPath,chargeDataPath,keytime):
    print "begin"
    wordOne = "userinfo"
    dataUser = search(userInfoDataPath, wordOne)
    wordTwo = "charge"
    datacharge = search(chargeDataPath, wordTwo)
    informationData = {}
    thirty = []
    today = ""
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
                        informationData.setdefault(sid,[[],[],[],[],[],[],[],[],[],[],[]])
                        informationData[sid][10].append(gid)
                    if int(line[cLvl]) >= 30:
                        thirty.append(uid)
                    if line[lastLogginIn][0:10] == ffff[len(ffff) - 1][0:10]:
                        informationData[sid][1].append(uid)
                    if line[new][0:10] == ffff[len(ffff) - 1][0:10]:
                        informationData[sid][2].append(uid)
                except Exception,e:
                    print temp


    for key in informationData:
        for subkey in range(len(informationData[key][2])):
            if informationData[key][2][subkey] in thirty:
                informationData[key][3].append(informationData[key][2][subkey])

    print "read charge"
    for file in datacharge:
        with open(file) as f:
            try:
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
                    informationData[sid][5].append(uid)

                    if uid in informationData[sid][2]:
                        informationData[sid][6].append(uid)
                        informationData[sid][7].append(pay)
                    if uid in informationData[sid][3]:
                        informationData[sid][8].append(uid)
                        informationData[sid][9].append(pay)
            except Exception,e:
                print Exception,e
                print x



    sortltvData = [(k, informationData[k]) for k in sorted(informationData.keys())]
    json_str = json.dumps(sortltvData)
    store("/data/historyQinmi/QinmiData_%s.json"%keytime, json_str)
    print "/data/historyQinmi/QinmiData_%s.json"%keytime+"ok"

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

def writeQinmi():
    filePath= search('/data/historyQinmi','2017-03-')
    # 写入数据
    print filePath
    print "write now"
    workbook = xlwt.Workbook(encoding='ascii')
    worksheet = workbook.add_sheet('qin')

    for t in range(len(title)):
        worksheet.write(0, t, title[t])
    count = 1
    for keyfile in filePath:
        print keyfile
        temp = keyfile.split('_')

        with open(keyfile) as f:
            sortltvData = json.load(f)
        for key in sortltvData:
            # [0.[大于30级的uid],1.[当天登入的uid],2.[当天新注册的UID],3.[滚服账号uid],4.[付费金额],5.[付费账号uid],
            # 6.[新增账号付费人数],7.[新增账号付费金额],8.[滚服新增付费人数],9.[滚服新增付费金额]]

            worksheet.write(count, 0, temp[1][0:10])  # 日期
            worksheet.write(count, 1, key[1][10])  # 大区
            worksheet.write(count, 2, key[0])  # 区号
            worksheet.write(count, 3, len(key[1][2]))  # 新注册
            worksheet.write(count, 4, len(key[1][3]))  # 滚服新账号数
            worksheet.write(count, 5, len(key[1][1]))  # 活跃账号数
            worksheet.write(count, 6, len(key[1][1]) - len(key[1][2]))  # 活跃老账号数
            worksheet.write(count, 7, sum(key[1][4]))  # 付费金额
            worksheet.write(count, 8, len(key[1][5]))  # 付费账号数

            if len(key[1][5]) <= 0 or len(key[1][1]) <= 0:
                worksheet.write(count, 9, 0)
            else:
                worksheet.write(count, 9, float(len(key[1][5])) / float(len(key[1][1])))  # 活跃付费率

            if sum(key[1][4]) <= 0 or len(key[1][2]) <= 0:
                worksheet.write(count, 10, 0)
            else:
                worksheet.write(count, 10, float(sum(key[1][4])) / float(len(key[1][2])))  # ARPU

            if sum(key[1][4]) <= 0 or len(key[1][5]) <= 0:
                worksheet.write(count, 11, 0)
            else:
                worksheet.write(count, 11, float(sum(key[1][4])) / float(len(key[1][5])))  # ARPPU

            if sum(key[1][4]) <= 0 or len(key[1][1]) <= 0:
                worksheet.write(count, 12, 0)
            else:
                worksheet.write(count, 12, float(sum(key[1][4])) / float(len(key[1][1])))  # 活跃ARPU

            worksheet.write(count, 13, len(key[1][6]))  # 新增账号付费人数
            worksheet.write(count, 14, sum(key[1][7]))  # 新增账号付费金额
            worksheet.write(count, 15, len(key[1][8]))  # 滚服新增付费人数
            worksheet.write(count, 16, sum(key[1][9]))  # 滚服新增付费金额

            if len(key[1][6]) <= 0 or len(key[1][2]) <= 0:
                worksheet.write(count, 17, 0)
            else:
                worksheet.write(count, 17, float(len(key[1][6])) / float(len(key[1][2])))  # 新增账号付费率
            if sum(key[1][7]) <= 0 or len(key[1][2]) <= 0:
                worksheet.write(count, 18, 0)
            else:
                worksheet.write(count, 18, float(sum(key[1][7])) / float(len(key[1][2])))  # 新增账号付费ARPU
            count += 1
    workbook.save("./data.xls")
    print "end", datetime.datetime.now()

if __name__ == '__main__':
    # today = datetime.date.today()
    # yesterday = today - datetime.timedelta(days=1)
    # strYesterday = yesterday.strftime('%Y-%m-%d')
    #get_user_event_info(strYesterday, "./s3-data/"+strYesterday)
    # csvPath = os.listdir('/data/s3-data/userinfo')
    # hisfile = os.listdir('/data/historyQinmi')
    # for key in csvPath:
    #     if key < "2017-03-01":
    #         continue
    #     if "QinmiData_%s.json"%key in hisfile:
    #         continue
    #     print "keystr",key
    #     infomation("/data/s3-data/userinfo/" + key,"/data/s3-data/charge/" + key, key)
    writeQinmi()
