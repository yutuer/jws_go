# -*- coding:utf-8 -*-
import json
import datetime
import boto
from boto import s3
import StringIO
import os
import pandas

def write_json(buffer, finleValue):
    buffer.writelines(finleValue)


def get_s3_list(daterange,prefixpath):
    prefix = prefixpath
    total_size = 0
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    ret = []
    if bucket:
        for k in bucket.list(prefix=prefix):
            if k.size <= 0:
                continue
            logsp = k.name.split('.')
            a = logsp[-4:-1]
            a.reverse()
            dt = ''.join(a)
            if dt == daterange:
                total_size += k.size
                ret.append('s3://prodlog/' + k.name)
                print('s3://prodlog/' + k.name, ''.join(a))
    print('total:%d' % (total_size / 1024.0 / 1024.0 / 1024.0))
    return ret

def readLog(s_path):
    with open(s_path,'r') as f:
        json_buffer = StringIO.StringIO()
        for line in f:
            try:
                dataLine = json.loads(line, encoding="utf-8")
            except Exception,e:
                print line
            if dataLine['type_name'] == "IAP":
                write_json(json_buffer,line)
            elif dataLine['type_name'] == "CorpLevelChg":
                write_json(json_buffer, line)
            elif dataLine['type_name'] == "Login":
                write_json(json_buffer, line)
            elif dataLine['type_name'] == "CreateProfile":
                write_json(json_buffer, line)
            elif dataLine['type_name'] == "FirstPay":
                write_json(json_buffer, line)
    return json_buffer

def upload_to_s3(json_buffer, region, bucket_name, s3_filepath):
    conn = boto.s3.connect_to_region(region)
    bucket = conn.get_bucket(bucket_name)
    full_key_name = s3_filepath
    k = bucket.new_key(full_key_name)
    k.set_contents_from_string(json_buffer.getvalue())


if __name__ == '__main__':
    begin = datetime.date(2016, 11, 11)
    end = datetime.date(2017, 6, 21)
    for i in range((end - begin).days + 1):
        temp = str(begin + datetime.timedelta(days=i))
        pp = temp.split('-')
        day = "".join(pp)
        print "day",day
        ret = get_s3_list(day,"logics3/")
        for file_path in ret:
            #downfile(file_path,"./temp")
            tempName = file_path.split("/")
            logcalName = tempName[len(tempName)-1]
            for i in range(len(tempName)):
                if tempName[i] == "logics3":
                    tempName[i] = "dataForSpark"
            newS3Name = "/".join(tempName[-3:])
            os.system("aws --region cn-north-1 s3 cp %s  ./clear/"%file_path)
            Gz = file_path.split(".")
            if Gz[len(Gz) - 1] == "gz":
                os.system("gunzip ./clear/%s"%logcalName)
                xx = logcalName.split('.')
                json_buffer = readLog("./clear/%s"%".".join(xx[:4]))
                ss = newS3Name.split(".")
                upload_to_s3(json_buffer, 'cn-north-1', "prodlog", ".".join(ss[:len(ss)-1]))
                print ".".join(ss[:len(ss)-1])
                os.system("rm -rf ./clear/%s"%".".join(xx[:4]))
            else:
                json_buffer = readLog("./clear/%s"%logcalName)
                upload_to_s3(json_buffer, 'cn-north-1', "prodlog", newS3Name)
                print newS3Name
                os.system("rm -rf ./clear/%s"%logcalName)