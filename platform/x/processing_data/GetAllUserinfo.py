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
            a = logsp[len(logsp)-1].split('.')[0]
            #suffix = logsp[len(logsp)-1].split('.')[1]

            if a == daterange[0] and str(logsp[2]) == 'userinfo' and logsp[len(logsp)-1].split('.')[1] == 'csv':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retUser.append(k.name)
            if a == daterange[0] and str(logsp[2]) == 'event' and logsp[len(logsp)-1].split('.')[1] == 'log':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retPay.append(k.name)
    print('total:%d'%(total_size/1024.0/1024.0/1024.0))
    # print retUser
    # print retPay

    payNum = 0
    userNum = 0
    while payNum < len(retPay):
        try:
            print "file Num",payNum
            downfile(retPay[payNum],daterange,daterange+'/'+retPay[payNum])
            print retPay[payNum],,"File Down Succes"
            payNum += 1
        except:
            print "except:",payNum
            print retPay[payNum],"File try again"

    while userNum < len(retUser):
        try:
            print "file Num",userNum
            downfile(retUser[payNum],daterange,daterange+'/'+retUser[userNum])
            print retUser[userNum],"File Down Succes"
            userNum += 1
        except:
            print "except:",userNum
            print retUser[userNum],"File try again"
# key event/2017-03-11/134_130134004019_event_2017-03-11.log
# key bilogs-csv/2017-03-11/134_130134004013_userinfo_2017-03-11.csv
def downfile(s3key,dPath,savePath):
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    dataPath = s3key.splint('/')

    if not os.path.exists(dPath+'/bilogs-csv/'+dataPath[1]):
        os.makedirs(dPath+'/bilogs-csv/'+dataPath[1])
    if not os.path.exists(dPath+'/event/'+dataPath[1]):
        os.makedirs(dPath+'/event/'+dataPath[1])

    print bucket
    if bucket:
        key = bucket.get_key(s3key)
        print key
        key.get_contents_to_filename(savePath)

if __name__ == '__main__':
    get_userinfo("/Users/tq/Documents")