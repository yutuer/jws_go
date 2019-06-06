import boto
from boto import s3
import StringIO

import pytz
import datetime


def upload_to_s3(region, bucket_name, s3_filepath):
    conn = boto.s3.connect_to_region(region)
    bucket = conn.get_bucket(bucket_name)
    full_key_name = s3_filepath
    k = bucket.new_key(full_key_name)
    k.set_contents_from_filename(csv_buffer.getvalue())
    return None

if __name__ == '__main__':
    now_utc = datetime.datetime.utcnow()
    local_tz = pytz.timezone('Asia/Shanghai')
    now_utc = pytz.utc.localize(now_utc)
    local_time = now_utc.astimezone(local_tz)
    current_date = local_time.strftime('%Y%m%d')
    target_path = 'devlog'
    target_path += '/%s' % current_date

    upload_to_s3('cn-north-1', "devlog", '/%s/test'%target_path).csv