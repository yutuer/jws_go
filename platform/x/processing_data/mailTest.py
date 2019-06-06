# -*- coding:utf-8 -*-
from __future__ import print_function
import numpy as np
import boto
from boto import s3
import StringIO
import sys, os
import pytz
import datetime

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


if __name__ == "__main__":
    today = datetime.date.today()
    yesterday = today - datetime.timedelta(days=1)
    strYesterday = yesterday.strftime('%Y-%m-%d')
    with open('/Users/tq/Documents/134_130134004019_charge_2017-03-15.log') as f:
        as_attachments(os.path.basename('/Users/tq/Documents/134_130134004019_charge_2017-03-15.log'), f)
    #livedata@taiyouxi.cn
        sys.exit(email_report('huquanqi@taiyouxi.cn', u'IFSG 数据留存计算 %s'%(strYesterday), u'''
    IFSG 数据留存计算
    测试数据效果
    %s
    '''%(strYesterday)))
