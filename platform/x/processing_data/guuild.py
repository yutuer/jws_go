# -*- coding: utf-8 -*-

from boto import s3
import os
import json

def get_logics(daterange):

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
            if len(logsp) < 2:
                continue
            print 'name',logsp
            temp = logsp[0].split('/')
            if temp[0] == 'logics3' and temp[1] == 'psid=4001':
                total_size += k.size
                print('s3://prodlog/' + k.name, ''.join(a))
                retUser.append(k.name)

    print('total:%d'%(total_size/1024.0/1024.0/1024.0))

    print "im here1"
    if not os.path.exists(daterange+'/logics/'):
        print "im here2"
        os.makedirs(daterange+'/logics/')

    #logics3/psid=4001/logics_shard4001.27.02.2017.gz
    payNum = 0
    userNum = 0
    while userNum < len(retUser):
        try:
            print "file Num",userNum
            finalPath = downfile(retUser[userNum],daterange)
            print retUser[userNum],"File Down Succes"
            userNum += 1
            temp = retUser[userNum].split('.')
            gz = os.listdir(daterange+'/logics/')
            gzpath = daterange+'/logics/'+ gz[0]
            if temp[len(temp)- 1] == 'gz':
                os.system('gunzip %s'%gzpath)
                filename = os.listdir(daterange+'/logics/')
                filenamePath = daterange+'/logics/'+filename[0]
                writeFile(filenamePath)
                os.system('rm -rf %s'%filenamePath)
            else:
                writeFile(finalPath)
                os.system(finalPath)

        except Exception,e:
            print Exception,e
            print "except:",userNum
            print retUser[userNum],"File try again"


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
        print (dPath + '/logics/' + name)
        key.get_contents_to_filename(dPath + '/logics/' + name)
    finalPath = dPath + '/logics/' + name
    return finalPath

def writeFile(file):
    x = open('/home/ec2-user/result4001/result4001.txt', 'a')
    with open(file) as f:
        for line in f:
            d = json.loads(line)
            if d['type_name'] == 'AddGuildInventory':
                x.write('\n'+line)
            elif d['type_name'] == 'AssignGuildInventory':
                x.write('\n'+line)
    x.close()




if __name__ == '__main__':
    get_logics('./')