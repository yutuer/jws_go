#!/usr/bin/python
# coding=gbk
# 用法：python diff_excel.py [file1] [file2]

from xlrd import open_workbook
import time
import os


# 输出整个Excel文件的内容
def print_workbook(wb):
    for s in wb.sheets():
        print "Sheet:", s.name
        for r in range(s.nrows):
            strRow = ""
            for c in s.row(r):
                strRow += ("\t" + c.value)
            print "ROW[" + str(r) + "]:", strRow

            # 把一行转化为一个字符串


def row_to_str(row):
    strRow = ""
    for c in row:
        strRow += ("\t" + c.value)
    return strRow;


# 打印diff结果报表
def print_report(report):
    for o in report:
        if isinstance(o, list):
            for i in o:
                print "\t" + i
        else:
            print o

            # diff两个Sheet

# 输出到TXT diff结果报表
def out_report(report,x,filename,sheetname):
    if len(report)>0:
        x.write("\n###################################################" + "\n")
        x.write("diff file: %s Sheet Name : %s \n" % (filename,sheetname.encode('utf8')))
        for o in report:
            if isinstance(o, list):
                for i in o:
                    x.write( "\n\t" + i.encode('utf8'))
            else:
                x.write("\n" + o.encode('utf8'))


def diff_sheet(sheet1, sheet2):
    nr1 = sheet1.nrows
    nr2 = sheet2.nrows
    nr = max(nr1, nr2)
    report = []
    for r in range(nr):
        row1 = None;
        row2 = None;
        if r < nr1:
            row1 = sheet1.row(r)
        if r < nr2:
            row2 = sheet2.row(r)
        if row1 is None:
            row1 = ""
        if row2 is None:
            row2 = ""

        diff = 0;  # 0:equal, 1: not equal, 2: row2 is more, 3: row2 is less
        if row1 == None and row2 != None:
            diff = 2
            report.append("+ROW[" + str(r + 1) + "]: " + row_to_str(row2))
        if row1 == None and row2 == None:
            diff = 0
        if row1 != None and row2 == None:
            diff = 3
            report.append("-ROW[" + str(r + 1) + "]: " + row_to_str(row1))
        if row1 != None and row2 != None:
            # diff the two rows
            reportRow = diff_row(row1, row2)
            if len(reportRow) > 0:
                report.append("#ROW[" + str(r + 1) + "]Befor: ")
                report.append("#ROW[" + str(r + 1) + "]After: ")
                report.append(reportRow)

    return report;


# diff两行
def diff_row(row1, row2):
    nc1 = len(row1)
    nc2 = len(row2)
    nc = max(nc1, nc2)
    report = []
    for c in range(nc):
        # try:
        ce1 = None;
        ce2 = None;
        if c < nc1:
            ce1 = row1[c]
        if c < nc2:
            ce2 = row2[c]

        if ce1 == None:
            ce1Value = ""
        elif isinstance(ce1.value,float):
            ce1Value = str(ce1.value)
        else:
            ce1Value = ce1.value

        if ce2 == None:
            ce2Value == ""
        elif isinstance(ce2.value,float):
            ce2Value = str(ce2.value)
        else:
            ce2Value = ce2.value
        diff = 0;  # 0:equal, 1: not equal, 2: row2 is more, 3: row2 is less
        if ce1 == None and ce2 != None:
            diff = 2
            report.append("+CELL[" + str(c + 1) + ": " + ce2Value)
        if ce1 == None and ce2 == None:
            diff = 0
        if ce1 != None and ce2 == None:
            diff = 3
            report.append("-CELL[" + str(c + 1) + ": " + ce1Value)
        if ce1 != None and ce2 != None:
            if ce1.value == ce2.value:
                diff = 0
            else:
                diff = 1
                report.append("#CELL[" + str(c + 1) + "]Befor: " + ce1Value)
                report.append("#CELL[" + str(c + 1) + "]After: " + ce2Value)
        # except Exception,e:
        #     print Exception,e
    return report

def search(path, word):
    name = []
    for filename in os.listdir(path):
        if word in filename:
            name.append(filename)
    return name


if __name__ == '__main__':
    # if len(sys.argv) < 3:
    #     exit()
    #
    # file1 = sys.argv[1]
    # file2 = sys.argv[2]
    newFile = []
    file1 = search("./befor", ".xlsx")
    file2 = search("./after", ".xlsx")

    for x in file2:
        if x not in file1:
            newFile.append(x)

    with open('./result/diff.txt', 'wt') as output:
        print time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(time.time()))
        output.write("DIff Time: " + time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(time.time())))
        for keyfile in file1:
            wb1 = open_workbook("./befor/"+keyfile)
            if os.path.exists("./after/"+keyfile):
                wb2 = open_workbook("./after/"+keyfile)
            else:
                output.write("\n###################################################" + "\n")
                output.write("file : %s not Exists!!! \n" % keyfile)
                continue
            wb1sheets = wb1.sheet_names()
            wb2sheets = wb2.sheet_names()
            newSheets = []
            for x in wb2sheets:
                if x not in wb1sheets:
                    newSheets.append(x)
            # diff两个文件的sheet
            for x in range(len(wb1sheets)):
                print ("###################################################")
                #output.write("\n###################################################" + "\n")
                if wb1sheets[x] not in wb2sheets:
                    output.write("\n###################################################" + "\n")
                    output.write("\nThere is no sheet : %s in after file : %s" %(wb1sheets[x],keyfile))
                    continue
                print "diff file %s Sheet Name %s : " %(keyfile,wb1sheets[x])
                #output.write("diff file: %s Sheet Name : %s \n" % (keyfile,wb1sheets[x].encode('utf8')))
                report = diff_sheet(wb1.sheet_by_name(wb1sheets[x]), wb2.sheet_by_name(wb1sheets[x]))
                # 打印diff结果
                print_report(report)
                out_report(report,output,keyfile,wb1sheets[x])
            if len(newSheets) >0:
                output.write("\n##This is new sheet in this Excel:\n")
                for x in newSheets:
                    output.write("\t: " + x)
        # 打印新增文件
        output.write("\n###################################################" + "\n")
        output.write("This is new Excel : \n")
        for x in newFile:
                output.write("\t" + x)
                # if len(report) == 0:
                #     print "                  There is no different !!!!"
                #     output.write("                  There is no different !!!! \n")
    output.close()