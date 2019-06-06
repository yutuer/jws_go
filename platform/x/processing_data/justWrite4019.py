# -*- coding: utf-8 -*-
import json
import os


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

def writeFile(file):
    x = open('/home/ec2-user/result4001/result4019.txt', 'a')
    with open(file) as f:
        for line in f:
            d = json.loads(line)
            if d['type_name'] == 'AddGuildInventory':
                x.write('\n'+line)
            elif d['type_name'] == 'AssignGuildInventory':
                x.write('\n'+line)
    print file,"ok"

    has.append(file)
    x.close()


if __name__ == '__main__':
        search('/data/s3-data/4019')
        print name
        for x in name:
            writeFile(x)
            print len(has)
