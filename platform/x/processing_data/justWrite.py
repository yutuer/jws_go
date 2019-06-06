# -*- coding: utf-8 -*-
import json
import os
import sys


name = []
has = []

def search(path):
    for filename in os.listdir(path):
        fp = os.path.join(path, filename)
        if os.path.isfile(fp):
            name.append(fp)
        elif os.path.isdir(fp):
            search(fp)
    return name

def writeFile(file,sid):
    x = open('/data/s3-data/result/result%s.txt'%sid, 'a')
    with open(file) as f:
        for line in f:
            try:
                d = json.loads(line)
                if d['type_name'] == 'AddGuildInventory':
                    x.write(line)
                elif d['type_name'] == 'AssignGuildInventory':
                    x.write(line)
            except Exception,e:
                print Exception,e
                print line
    print file,"ok"

    has.append(file)
    x.close()


if __name__ == '__main__':
        sid = sys.argv[1]
        if not os.path.exists('/data/s3-data/%s'%sid):
            os.makedirs('/data/s3-data/%s'%sid)
        os.system('aws --region cn-north-1 s3 cp s3://prodlog/logics3/psid=%s /data/s3-data/%s --recursive'%(sid,sid))
        os.system('sh /data/s3-data/xxx.sh %s'%sid)
        search('/data/s3-data/%s'%sid)
        print name
        for x in name:
            writeFile(x,sid)
            print len(has)
