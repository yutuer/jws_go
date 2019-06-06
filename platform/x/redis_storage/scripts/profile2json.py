#!/usr/bin/python
# -*- coding: utf-8 -*-
'''
转化profile前缀的玩家存档为Json
然后转化json用于辅助分析玩家profile存档下每个字段，
随着玩家战队等级的尺寸变化情况
'''
import json
import re
import fileinput
import sys

deepclean = re.compile(r'\\+"')
def processline(line):
    if line.startswith('profile'):
        s = line.split(' ', 1)
        accountid = s[0]
        profile = s[1]
        p = deepclean.sub('"', profile)
        p = p.replace('"{', '{')
        p = p.replace('}"', '}')
        p = p.replace(']"', ']')
        p = p.replace('"[', '[')
        try:
            d = json.loads(p)
            lvl = d['corp']['lv']
            for k in d:
                sys.stdout.write('%s,%d,%s,%d\n'%(accountid, lvl, k, len(json.dumps(d[k]))))
        except:
            sys.stderr.write('%s %s'%(accountid,p))

for line in fileinput.input('-'):
    processline(line)

