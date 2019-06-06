import boto
from boto import s3
import os

def get_userinfo(daterange):
    daterange = daterange.split(',')
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
            print logsp
            a = logsp[len(logsp)-1].split('.')[0]
            #suffix = logsp[len(logsp)-1].split('.')[1]
            b = a.split('-')
            dt = ''.join(b)
            if a == daterange[0] and str(logsp[2]) == 'userinfo' and logsp[len(logsp)-1].split('.')[1] == 'csv':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retUser.append(k.name)
            if a == daterange[0] and str(logsp[2]) == 'event' and logsp[len(logsp)-1].split('.')[1] == 'log':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retPay.append(k.name)
    print('total:%d'%(total_size/1024.0/1024.0/1024.0))
    print retUser
    print retPay
    if not os.path.exists(daterange[1]+'/bilogs-csv/'+daterange[0]):
        os.makedirs(daterange[1]+'/bilogs-csv/'+daterange[0])
    if not os.path.exists(daterange[1]+'/event/'+daterange[0]):
        os.makedirs(daterange[1]+'/event/'+daterange[0])

    payNum = 0
    userNum = 0
    while payNum < len(retPay):
        try:
            print "file Num",payNum
            downfile(retPay[payNum],daterange[1]+'/'+retPay[payNum])
            print retPay[payNum],"File Down Succes"
            payNum += 1
        except:
            print "except:",payNum
            print retPay[payNum],"File try again"

    while userNum < len(retUser):
        try:
            print "file Num",userNum
            downfile(retUser[payNum],daterange[1]+'/'+retUser[userNum])
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

def test():
    with open("/Users/tq/Documents/134_130134004019_charge_2017-03-15.log") as f:
        for line in f:
            x =line.split('$$')
            print x
            print x[1],x[8],x[11][0:10]
            s = x[1].split(":")
            print s[0]

def test1():
    with open("/Users/tq/data_analysis/event/134_110134002001_event_2017-03-08.log") as f:
        for line in f:
            x =line.split('$$')
            print x

if __name__ == '__main__':
    #get_userinfo("2017-03-11,/Users/tq/Documents")
    test1()