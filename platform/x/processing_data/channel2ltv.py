# -*- coding: utf-8 -*-

import json
import os
import datetime
import xlwt


def acid2CreatTime(chargeDataPath):
    wordTwo = "charge"
    datacharge = search(chargeDataPath)
    ltvData = {}
    ltvData200 = {}
    ltvData201 = {}
    ltvData202 = {}
    ltvData203 = {}
    with open("/home/ec2-user/ltvtest/uid2Time_200.json") as f:
        data200 = json.load(f)
    with open("/home/ec2-user/ltvtest/uid2Time_201.json") as f:
        data201 = json.load(f)
    with open("/home/ec2-user/ltvtest/uid2Time_202.json") as f:
        data202 = json.load(f)
    with open("/home/ec2-user/ltvtest/uid2Time_203.json") as f:
        data203 = json.load(f)
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
                        if x[3] not in ltvData:
                            ltvData.setdefault(x[3],{})
                        if data200.has_key(uid):
                            creatTime = data200[uid]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData[x[3]].keys():
                            ltvData[x[3]].setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData[x[3]][creatTime].keys():
                            ltvData[x[3]][creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData[x[3]][creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData[x[3]][creatTime][x[11][0:10]].append(float(x[8]))
                    if sid[0] == "201":
                        if x[3] not in ltvData:
                            ltvData.setdefault(x[3],{})
                        if data201.has_key(uid):
                            creatTime = data201[uid]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData[x[3]].keys():
                            ltvData[x[3]].setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData[x[3]][creatTime].keys():
                            ltvData[x[3]][creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData[x[3]][creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData[x[3]][creatTime][x[11][0:10]].append(float(x[8]))
                    if sid[0] == "202":
                        if x[3] not in ltvData:
                            ltvData.setdefault(x[3],{})
                        if data202.has_key(uid):
                            creatTime = data202[uid]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData[x[3]].keys():
                            ltvData[x[3]].setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData[x[3]][creatTime].keys():
                            ltvData[x[3]][creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData[x[3]][creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData[x[3]][creatTime][x[11][0:10]].append(float(x[8]))

                    if sid[0] == "203":
                        if x[3] not in ltvData:
                            ltvData.setdefault(x[3],{})
                        if data203.has_key(uid):
                            creatTime = data203[uid]
                        else:
                            continue
                        if len(creatTime) > 15:
                            print creatTime,line
                            continue
                        if creatTime not in ltvData[x[3]].keys():
                            ltvData[x[3]].setdefault(creatTime,{})
                        if x[11][0:10] == '2016-10-26':
                            continue
                        if x[11][0:10] not in ltvData[x[3]][creatTime].keys():
                            ltvData[x[3]][creatTime].setdefault(x[11][0:10],[])
                        if x[8].isdigit():
                            ltvData[x[3]][creatTime][x[11][0:10]].append(int(x[8]))
                        else:
                            ltvData[x[3]][creatTime][x[11][0:10]].append(float(x[8]))
                except Exception,e:
                    print Exception,e
                    print "linlin",line

    json_str = json.dumps(ltvData)

    store("ltvChannelData.json",json_str)

    #写入数据
    print "write now"
    workbook = xlwt.Workbook(encoding='ascii')

    for ch in ltvData:
        worksheet = workbook.add_sheet(ch)

        sortltvData = [(k, ltvData[ch][k]) for k in sorted(ltvData[ch].keys())]
        t = 0
        for key in sortltvData:
            worksheet.write(0,t+2,key[0])
            t += 1

        count = 0
        subcount = 1
        peoples = 0
        paySum = 0
        try:
            for key in sortltvData:
                if key[0] < "2016-11-10":
                    continue
                if len(key[0]) > 10:
                    continue
                count +=1
                worksheet.write(count, 0, key[0])

                for acidkey in data200:
                    if data200[acidkey] == key[0]:
                        peoples += 1
                for acidkey in data201:
                    if data201[acidkey] == key[0]:
                        peoples += 1
                for acidkey in data202:
                    if data202[acidkey] == key[0]:
                        peoples += 1
                for acidkey in data203:
                    if data203[acidkey] == key[0]:
                        peoples += 1


                worksheet.write(count, 1, peoples)

                sortkey = [(k,key[1][k]) for k in sorted(key[1].keys())]

                for subkey in sortkey:
                    if subkey[0] < "2016-11-10":
                        continue
                    if len(subkey[0]) > 10:
                        continue
                    subcount += 1
                    paySum += sum(subkey[1])

                    day = getRelativedays(key[0],subkey[0])
                    print "dayday",day
                    #worksheet200.write(count,subcount,sum(subkey[1]))
                    worksheet.write(count, day + 2, float(paySum) / float(peoples))
                paySum = 0
                subcount = 1
                peoples = 0
        except Exception, e:
            print "200 write wrong",Exception, e
            print key[0],"####",key[1]
            print "201day", day, key[0], subkey[0],subkey

    workbook.save("/home/ec2-user/channelLtv/lty.xls")
    print "end",datetime.datetime.now()

name = []
def search(path):
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

if __name__ == '__main__':
    #get_user_event_info("/home/ec2-user/ltv")
    acid2CreatTime("/data/s3-data/charge")