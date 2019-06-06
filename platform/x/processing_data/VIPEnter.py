# -*- coding: utf-8 -*-
import os
import csv
import xlwt
import sys
from boto import s3
import datetime
import MySQLdb

def VipEnter(userInfoDataPath,eventDataPath,makePath,dateDate):
    # oneDataPath = "/Users/tq/data_analysis/data_analysis_project/userinfo"
    # towDataPath = "/Users/tq/data_analysis/data_analysis_project/event"

    title = ["Date","Gid","VIPLvl","LoggingIn","HCNum","CostHCNum","GS","CorpLvl","HCBuy","HcGive"]
    wordOne = "userinfo"
    wordTwo = "event"
    dataUser = search(userInfoDataPath,wordOne)
    dataCurrency = search(eventDataPath,wordTwo)
    print len(dataUser),len(dataCurrency)
    VIPEnterData = {}
    VIPEnterDataBysid = {}
    allpeople = 0
    for file in dataUser:
        xxxx = file.split('/')
        ffff = xxxx[len(xxxx) - 1].split('_')
        print ffff
        with open(file) as f:
            reader = csv.reader(f)
            for line in reader:
                if line[0] == "accountId":
                    vip = line.index("VIP等级")
                    cLvl = line.index("战队等级")
                    gs = line.index("战队总战力")
                    hc = line.index("硬通数")
                    lastLogginIn = line.index("最后登陆日")
                    continue
                try:
                    if line[lastLogginIn][0:10] == ffff[len(ffff)-1][0:10]:
                        allpeople +=1
                        gid = line[0].split(':')
                        sid = gid[1]
                        if gid[0] not in VIPEnterData.keys():
                            VIPEnterData.setdefault(gid[0],{})

                        if sid not in VIPEnterDataBysid.keys():
                            VIPEnterDataBysid.setdefault(sid,{})
                        vipLvl = line[vip]
                        if vipLvl == "":
                            continue
                        if vipLvl not in VIPEnterData[gid[0]].keys():
                            VIPEnterData[gid[0]].setdefault(vipLvl,[[],[],[],[],[],[],[]])

                        if vipLvl not in VIPEnterDataBysid[sid].keys():
                            VIPEnterDataBysid[sid].setdefault(vipLvl,[[],[],[],[],[],[],[]])


                        VIPEnterData[gid[0]][vipLvl][0].append(gid[2])
                        VIPEnterData[gid[0]][vipLvl][1].append(int(line[cLvl]))
                        if int(line[cLvl]) >100:
                            print file,cLvl
                        VIPEnterData[gid[0]][vipLvl][2].append(int(line[gs]))
                        VIPEnterData[gid[0]][vipLvl][3].append(int(line[hc]))

                        VIPEnterDataBysid[sid][vipLvl][0].append(gid[2])
                        VIPEnterDataBysid[sid][vipLvl][1].append(int(line[cLvl]))
                        VIPEnterDataBysid[sid][vipLvl][2].append(int(line[gs]))
                        VIPEnterDataBysid[sid][vipLvl][3].append(int(line[hc]))

                except Exception, e:
                    print Exception,e,file,gid

    print "Userinfo read ok"
    for file in dataCurrency:
        with open(file) as f:
            print file
            for line in f:
                x = line.split('$$')
                try:
                    if x[1] == "REMOVE_MONEY" and x[7] == "VI_HC":
                        accountId = x[3].split(':')
                        for ss in VIPEnterData[accountId[0]]:
                            if accountId[1] in VIPEnterData[accountId[0]][ss][0]:
                                costHc = int(x[5])-int(x[6])
                                VIPEnterData[accountId[0]][ss][4].append(costHc)
                        sidtmp = x[4].split(':')
                        sid = sidtmp[1]
                        for xx in VIPEnterDataBysid[sid]:
                            if accountId[1] in VIPEnterDataBysid[sid][xx][0]:
                                costHc = int(x[5]) - int(x[6])
                                VIPEnterDataBysid[sid][xx][4].append(costHc)
                    if x[1] == "GET_MONEY" and x[7] == "VI_HC_Give":
                        accountId = x[3].split(':')
                        for ss in VIPEnterData[accountId[0]]:
                            if accountId[1] in VIPEnterData[accountId[0]][ss][0]:
                                costHc = int(x[6])-int(x[5])
                                VIPEnterData[accountId[0]][ss][5].append(costHc)

                        sidtmp = x[4].split(':')
                        sid = sidtmp[1]
                        for xx in VIPEnterDataBysid[sid]:
                            if accountId[1] in VIPEnterDataBysid[sid][xx][0]:
                                costHc = int(x[6]) - int(x[5])
                                VIPEnterDataBysid[sid][xx][5].append(costHc)

                    if x[1] == "GET_MONEY" and x[7] == "VI_HC_Buy":
                        accountId = x[3].split(':')
                        for ss in VIPEnterData[accountId[0]]:
                            if accountId[1] in VIPEnterData[accountId[0]][ss][0]:
                                costHc = int(x[6])-int(x[5])
                                VIPEnterData[accountId[0]][ss][6].append(costHc)

                        sidtmp = x[4].split(':')
                        sid = sidtmp[1]
                        for xx in VIPEnterDataBysid[sid]:
                            if accountId[1] in VIPEnterDataBysid[sid][xx][0]:
                                costHc = int(x[6]) - int(x[5])
                                VIPEnterDataBysid[sid][xx][6].append(costHc)
                except Exception, e:
                    print Exception,repr(e),file,x

    database = MySQLdb.connect(host="54.223.192.252", user="test1", passwd="QmPhaQ8hYsxx", db="test")
    cursor = database.cursor()

    # 创建插入SQL语句
    queryByGid = """INSERT INTO VipEnterByGid (Date,Gid,VIPLvl,LoggingIn,HCNum,CostHcNum,GS,CorpLvl,HCBuy,HCGive)
     VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"""

    queryBySid = """INSERT INTO VipEnterBySid (Date,Sid,VIPLvl,LoggingIn,HCNum,CostHcNum,GS,CorpLvl,HCBuy,HCGive)
     VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"""
    print "paylog read ok"
    #写入数据
    workbook = xlwt.Workbook(encoding='ascii')
    worksheet = workbook.add_sheet('VIPEnter')
    worksheetsid = workbook.add_sheet('VIPEnterBySid')
    for titleKey in range(len(title)):
        worksheet.write(0,titleKey,title[titleKey])
        worksheetsid.write(0, titleKey, title[titleKey])
    count = 0
    for gidKey in VIPEnterData:
        for vipKey in VIPEnterData[gidKey]:
            try:
                count += 1
                worksheet.write(count, 0, dateDate)
                Date = dateDate
                worksheet.write(count, 1, gidKey)
                Gid = gidKey
                worksheet.write(count, 2, vipKey)
                VIPLvl = vipKey
                worksheet.write(count, 3, len(VIPEnterData[gidKey][vipKey][0]))
                LoggingIn = len(VIPEnterData[gidKey][vipKey][0])
                worksheet.write(count, 4, sum(VIPEnterData[gidKey][vipKey][3]))
                HCNum = sum(VIPEnterData[gidKey][vipKey][3])
                worksheet.write(count, 5, sum(VIPEnterData[gidKey][vipKey][4]))
                CostHcNum = sum(VIPEnterData[gidKey][vipKey][4])
                worksheet.write(count, 6, sum(VIPEnterData[gidKey][vipKey][2])/len(VIPEnterData[gidKey][vipKey][2]))
                GS = sum(VIPEnterData[gidKey][vipKey][2])/len(VIPEnterData[gidKey][vipKey][2])
                worksheet.write(count, 7, sum(VIPEnterData[gidKey][vipKey][1])/len(VIPEnterData[gidKey][vipKey][1]))
                CorpLvl = sum(VIPEnterData[gidKey][vipKey][1])/len(VIPEnterData[gidKey][vipKey][1])
                worksheet.write(count, 8, sum(VIPEnterData[gidKey][vipKey][5]))
                HCBuy = sum(VIPEnterData[gidKey][vipKey][5])
                worksheet.write(count, 9, sum(VIPEnterData[gidKey][vipKey][6]))
                HCGive = sum(VIPEnterData[gidKey][vipKey][6])
                values = (Date,Gid,VIPLvl,LoggingIn,HCNum,CostHcNum,GS,CorpLvl,HCBuy,HCGive)

                # 执行sql语句
                cursor.execute(queryByGid, values)
            except Exception, e:
                print Exception,e,gidKey,vipKey
    count1 = 0
    for sidkey in VIPEnterDataBysid:
        for vipKey in VIPEnterDataBysid[sidkey]:
            try:
                count1 += 1
                worksheetsid.write(count1, 0, dateDate)
                Date = dateDate
                worksheetsid.write(count1, 1, sidkey)
                Sid = sidkey
                worksheetsid.write(count1, 2, vipKey)
                VIPLvl = vipKey
                worksheetsid.write(count1, 3, len(VIPEnterDataBysid[sidkey][vipKey][0]))
                LoggingIn = len(VIPEnterDataBysid[sidkey][vipKey][0])

                worksheetsid.write(count1, 4, sum(VIPEnterDataBysid[sidkey][vipKey][3]))
                HCNum = sum(VIPEnterDataBysid[sidkey][vipKey][3])

                worksheetsid.write(count1, 5, sum(VIPEnterDataBysid[sidkey][vipKey][4]))
                CostHcNum = sum(VIPEnterDataBysid[sidkey][vipKey][4])

                worksheetsid.write(count1, 6, sum(VIPEnterDataBysid[sidkey][vipKey][2]) / len(VIPEnterDataBysid[sidkey][vipKey][2]))
                GS = sum(VIPEnterDataBysid[sidkey][vipKey][2]) / len(VIPEnterDataBysid[sidkey][vipKey][2])

                worksheetsid.write(count1, 7, sum(VIPEnterDataBysid[sidkey][vipKey][1]) / len(VIPEnterDataBysid[sidkey][vipKey][1]))
                CorpLvl = sum(VIPEnterDataBysid[sidkey][vipKey][1]) / len(VIPEnterDataBysid[sidkey][vipKey][1])

                worksheetsid.write(count1, 8, sum(VIPEnterDataBysid[sidkey][vipKey][5]))
                HCBuy = sum(VIPEnterData[gidKey][vipKey][5])

                worksheetsid.write(count1, 9, sum(VIPEnterDataBysid[sidkey][vipKey][6]))
                HCGive = sum(VIPEnterData[gidKey][vipKey][6])

                values = (Date, Sid, VIPLvl, LoggingIn, HCNum, CostHcNum, GS, CorpLvl, HCBuy, HCGive)

                # 执行sql语句
                cursor.execute(queryBySid, values)
            except Exception,e:
                print Exception,e,sidkey,vipKey

    # 关闭游标
    cursor.close()

    # 提交
    database.commit()

    # 关闭数据库连接
    database.close()

    workbook.save(makePath+"/VIPEnter_"+dateDate+".xls")
    print "allpeopele",allpeople
    print "all write ok"

def search(path, word):
    name = []
    for filename in os.listdir(path):
        fp = os.path.join(path, filename)
        if os.path.isfile(fp) and word in filename:
            name.append(fp)
        elif os.path.isdir(fp):
            search(fp, word)
    return name

def get_userinfo(data, datePath):
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

            if a == data and str(logsp[2]) == 'userinfo' and logsp[len(logsp)-1].split('.')[1] == 'csv':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retUser.append(k.name)
            if a == data and str(logsp[2]) == 'event' and logsp[len(logsp)-1].split('.')[1] == 'log':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retPay.append(k.name)
    print('total:%d'%(total_size/1024.0/1024.0/1024.0))
    print type(data)
    if not os.path.exists(datePath+ '/bilogs-csv/'+str(data)):
        os.makedirs(datePath + '/bilogs-csv/' + str(data))
    if not os.path.exists(datePath+ '/event/'+str(data)):
        os.makedirs(datePath + '/event/' + str(data))

    payNum = 0
    userNum = 0
    while payNum < len(retPay):
        try:
            print "file Num",payNum
            downfile(retPay[payNum], datePath + '/' + retPay[payNum])
            print retPay[payNum],"File Down Succes"
            payNum += 1
        except:
            print "except:",payNum
            print retPay[payNum],"File try again"

    while userNum < len(retUser):
        try:
            print "file Num",userNum
            downfile(retUser[userNum], datePath + '/' + retUser[userNum])
            print retUser[userNum],"File Down Succes"
            userNum += 1
        except:
            print "except:",userNum
            print retUser[userNum],"File try again"

def downfile(s3key,path):
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    print bucket
    if bucket:
        key = bucket.get_key(s3key)
        print key
        key.get_contents_to_filename(path)


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
    print sys.argv
    # today = datetime.date.today()
    # yesterday = today - datetime.timedelta(days=1)
    # strYesterday = yesterday.strftime('%Y-%m-%d')
    begin = datetime.date(2016,12,1)
    end = datetime.date(2017,5,12)
    output = open('log.txt','a')
    if len(sys.argv) < 2:
        print 'No action specified.'
        sys.exit()
    if sys.argv[1].startswith('-'):
        option = sys.argv[1][1:]
        # fetch sys.argv[1] but without the first two characters
        if option == 'version':
            print 'Version 1.0'
        elif option == 'help':
            print '''
    This program will make VipEnter data.
    Options include:
      -version : Prints the version number
      -help    : Display this help
      -f       : Please Give me path where you want to down'''
        elif option == 'f':
            if len(sys.argv)< 3:
                print "Give me right two path userinfo.csv path eventlog path and make path"
                sys.exit()
            for i in range((end - begin).days + 1):
                strYesterday = str(begin + datetime.timedelta(days=i))
                print strYesterday

                get_userinfo(strYesterday,sys.argv[2])
                print "file download ok"
                dataDate = strYesterday
                VipEnter(sys.argv[2]+'/bilogs-csv/'+strYesterday,sys.argv[2]+'/event/'+strYesterday,sys.argv[2],dataDate)
                output.write(strYesterday)
                # os.popen('rm -rf /home/ec2-user/processing_data/bilogs-csv')
                # os.popen('rm -rf /home/ec2-user/processing_data/event')
                # with open(sys.argv[2] + "/VIPEnter.xls") as f:
                #     as_attachments(os.path.basename(sys.argv[2] + "/VIPEnter.xls"), f)
                #     # livedata@taiyouxi.cn
                #     sys.exit(email_report('huquanqi@taiyouxi.cn', u'VIP 数据留存计算 %s' % (strYesterday), u'''
                #    VIP 数据留存计算
                #    测试数据效果
                #    %s
                #    ''' % (strYesterday)))
            output.close()
        else:
            print 'Unknown option.'
        sys.exit()