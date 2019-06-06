# -*- coding: utf-8 -*-

import json
import os
import datetime
import xlwt


def acid2CreatTime(chargeDataPath):
    wordTwo = "charge"
    datacharge = search(chargeDataPath,wordTwo)
    ltvData200 = {}
    ltvData201 = {}
    ltvData202 = {}
    ltvData203 = {}
    with open("/home/ec2-user/acid2Time_200.json") as f:
        data200 = json.load(f)
    with open("/home/ec2-user/acid2Time_201.json") as f:
        data201 = json.load(f)
    with open("/home/ec2-user/acid2Time_202.json") as f:
        data202 = json.load(f)
    with open("/home/ec2-user/acid2Time_203.json") as f:
        data203 = json.load(f)

    print "### READ ok"
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
                        if data201.has_key(x[4]):
                            creatTime = data201[x[4]]
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
                        if data202.has_key(x[4]):
                            creatTime = data202[x[4]]
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
                        if data203.has_key(x[4]):
                            creatTime = data203[x[4]]
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
    subcount = 0
    try:
        for key in sortltvData200:
            count +=1
            worksheet200.write(count, subcount, key[0])

            sortkey = [(k,key[1][k]) for k in sorted(key[1].keys())]
            for subkey in sortkey:
                subcount += 1
                worksheet200.write(count,title.index(subkey[0])+1,sum(subkey[1]))
            subcount = 0
    except Exception, e:
        print "200 write wrong",Exception, e
        print key[0],"####",key[1]

    count1 = 0
    subcount1 = 0

    for key in sortltvData201:
        count1 +=1
        worksheet201.write(count1, subcount1, key[0])

        sortkey = [(k, key[1][k]) for k in sorted(key[1].keys())]
        for subkey in sortkey:
            subcount1 += 1
            worksheet201.write(count1, title.index(subkey[0])+1, sum(subkey[1]))
        subcount1 = 0

    count2 = 0
    subcount2 = 0
    for key in sortltvData202:
        count2 +=1
        worksheet202.write(count2, subcount2, key[0])

        sortkey = [(k, key[1][k]) for k in sorted(key[1].keys())]
        for subkey in sortkey:
            subcount2 += 1
            worksheet202.write(count2, title.index(subkey[0])+1, sum(subkey[1]))
        subcount2 = 0

    count3 = 0
    subcount3 = 0
    for key in sortltvData203:
        count3 +=1
        worksheet203.write(count3, subcount3, key[0])
        sortkey = [(k, key[1][k]) for k in sorted(key[1].keys())]
        for subkey in sortkey:
            subcount3 += 1
            worksheet203.write(count3, title.index(subkey[0])+1, sum(subkey[1]))
        subcount3 = 0


    workbook.save("/home/ec2-user/ltyTest.xls")
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

if __name__ == '__main__':
    #get_user_event_info("/home/ec2-user/ltv")
    acid2CreatTime("/home/ec2-user/s3-ltv/charge")