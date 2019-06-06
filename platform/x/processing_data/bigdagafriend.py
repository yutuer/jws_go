from boto import s3
from pyspark.sql import SparkSession
from pyspark import SparkContext, SparkConf

def getFalesChannel():
    false_channel = [
    "130134201",
    "110134101106.0",
    "13013412700",
    "13013419400",
    "130134164",
    "47",
    "13013415800",
    "1301343257",
    "13013413400",
    "131",
    "142",
    "145",
    "180",
    "188",
    "196",
    "204",
    "210",
    "213",
    "215",
    "216",
    "223",
    "234",
    "250",
    "259",
    "262",
    "263",
    "265",
    "292",
    "296",
    "298",
    "306",
    "309",
    "313",
    "1301341500",
    "1301341232",
    "13013416000",
    "13013411000",
    "1301344600",
    "1301345600",
    "13013412000",
    "1301341272",
    "13013415900",
    "130134137",
    "1301371335",
    "130134126",
    "1301343900",
    "13013415200",
    "13013423200",
    "1301343234",
    "130134312108",
    "1301341300",
    "130134312173",
    "5000",
    "130134159",
    "1301343800",
    "1301345400",
    "1301341",
    "130134144",
    "13013443300",
    "268",
    "350",
    "389"
]
    return false_channel

def getTrueChannel():
    true_channel = [
        "130134000201",
        "110134101106",
        "130134012700",
        "130134019400",
        "130134000164",
        "47",
        "130134015800",
        "130134003257",
        "130134013400",
        "131",
        "142",
        "145",
        "180",
        "188",
        "196",
        "204",
        "210",
        "213",
        "215",
        "216",
        "223",
        "234",
        "250",
        "259",
        "262",
        "263",
        "265",
        "292",
        "296",
        "298",
        "306",
        "309",
        "313",
        "130134001500",
        "130134001232",
        "130134016000",
        "130134011000",
        "130134004600",
        "130134005600",
        "130134012000",
        "130134001272",
        "130134015900",
        "130134000137",
        "130134001335",
        "130134000126",
        "130134003900",
        "130134015200",
        "130134023200",
        "130134003234",
        "130134312108",
        "130134001300",
        "130134312173",
        "5000",
        "130134000159",
        "130134003800",
        "130134005400",
        "130134000001",
        "130134000144",
        "130134043300",
        "268",
        "350",
        "389"
    ]
    return true_channel

def getDBDriver(tableName):
    return spark.read.format("jdbc").option("url", url).option("driver", "com.mysql.jdbc.Driver").option(
            "dbtable", tableName).option("user", properties['user']).option("password", properties['password']).load()


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

def get_s3_tempdata(dataName):
    prefix = "uidcreattime-csv/"
    total_size = 0
    REGION = "cn-north-1"
    conn = s3.connect_to_region(REGION)
    bucket = conn.lookup('prodlog')
    ret = []
    if bucket:
        for k in bucket.list(prefix=prefix):
            if k.size <= 0:
                continue
            logsp = k.name.split('/')
            a = logsp[2]
            csvName = a.split('.')[0]
            if csvName == dataName:
                total_size += k.size
                ret.append('s3://prodlog/' + k.name)
                print('s3://prodlog/' + k.name, ''.join(a))
    print('total:%d' % (total_size / 1024.0 / 1024.0 / 1024.0))
    return ret

def min_time(x, y):
    if int(x.lt) > int(y.lt):
        return y
    else:
        return x

def upload_to_s3(pandasdf, region, bucket_name, s3_filepath):
    for column in pandasdf.columns:
        for idx in pandasdf[column].index:
            x = pandasdf.get_value(idx, column)
            try:
                x = unicode(x.encode('utf-8', 'ignore'), errors='ignore') if type(x) == unicode else unicode(str(x),
                                                                                                             errors='ignore')
                pandasdf.set_value(idx, column, x)
            except Exception:
                print
                'encoding error: {0} {1}'.format(idx, column)
                pandasdf.set_value(idx, column, '')
                continue
    csv_buffer = StringIO.StringIO()
    pandasdf.to_csv(csv_buffer)
    conn = boto.s3.connect_to_region(region)
    bucket = conn.get_bucket(bucket_name)
    full_key_name = s3_filepath
    k = bucket.new_key(full_key_name)
    k.set_contents_from_string(csv_buffer.getvalue())
    return None
