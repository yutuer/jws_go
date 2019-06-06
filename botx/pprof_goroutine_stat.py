#coding:utf-8
## 用来对curl http://127.0.0.1:6061/debug/pprof/goroutine?debug=2 > gor.txt产生的结果进行统计
## 统计相同调用方法的goroutine的数量，并按字母顺序排列
import re
import sys

def fun():
    kv = {}
    patHeader = re.compile(r'goroutine \d.*')
    with open(sys.argv[1], "r") as f:
        isStep1 = False
        isStep2 = False
        k = ""
        lastLine = ""
        count = 0
        for line in f.readlines():
            if isStep1:
                match = re.compile(r'.*\(').match(line)
                if match:
                    k = match.group()
                    k = k[:len(k) - 1]

                    count += 1
                else:
                    print "!!!!!! err not find after [goroutine]: " + line
                isStep1 = False
                isStep2 = True
            elif isStep2 and len(line) <= 1:
                match = re.compile(r'.* \+').match(lastLine)
                if match:
                    tmp = match.group()
                    tmp = tmp[:len(tmp) - 1]
                    k += " - " + tmp.lstrip()

                    if kv.has_key(k):
                        kv[k] = kv[k] + 1
                    else:
                        kv[k] = 1
                else:
                    print "!!!!!! err not find after [created by]: " + lastLine

                isStep2 = False
            else:
                match = patHeader.match(line)
                if match:
                    if isStep2:
                        print "!!!!!! err not find [created by] but find [goroutine] again: " + lastLine + str(len(lastLine))
                        isStep2 = False
                    isStep1 = True
            lastLine = line

    ks = kv.keys()
    ks.sort()
    print "goroutine sum: " + str(count)
    for k in ks:
        print k + "  " + str(kv[k])

fun()