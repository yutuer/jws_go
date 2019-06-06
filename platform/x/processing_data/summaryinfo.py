# -*- coding: utf-8 -*-
import xlwt
import xlrd
import csv
import os
import sys

def Summary_data(activities_dir_name):

    summaryInfo = {}
    title = ["Date","DailyCreat","DailyEnter","DailyActivies","PayRMB","PayUserNum","DailyNewPay","PayPercent","ARPU","ARPPU"]
    activitiesData = []  #玩家activities总数据
    activitiesDate = []  #日期

    activitiesName = "/daily_activities.csv"
    payName = "pay/pay_new.xlsx"
    fpdName = "pay/pay_fpd.xlsx"


    pay_new_data = xlrd.open_workbook(activities_dir_name+payName)
    pay_fpd_data = xlrd.open_workbook(activities_dir_name+fpdName)
    table = pay_new_data.sheet_by_name(u'all_by_date')
    tableFpd = pay_fpd_data.sheet_by_name(u'FPD_byGidChFirst')
    rangeDir = os.listdir(activities_dir_name)
    for x in rangeDir:
        if x == "pay":
            continue
        tempActiveData = {}
        with open(activities_dir_name + x + activitiesName,'rb') as f:
            reader = csv.reader(f)
            for row in reader:
                    if reader.line_num == 1:
                        continue
                    if row[1] not in activitiesDate:
                        activitiesDate.append(row[1])
                    newActivitiesNum = int(row[2])
                    todayActivitiesNum = int(row[3])
                    tempActiveData.setdefault(row[1],[newActivitiesNum,todayActivitiesNum])

            f.close()
        activitiesData.append(tempActiveData)

    for date in activitiesDate:
        daily_create = 0
        daily_enter = 0
        for x in activitiesData:
            daily_create += x[date][0]
            daily_enter += x[date][1]
        daily_activies = daily_enter - daily_create
        summaryInfo.setdefault(date,{'DailyCreat':daily_create,
                                     'DailyEnter':daily_enter,
                                     'DailyActivies':daily_activies})

    #print summaryInfo
    for rownum in range(table.nrows):
        if table.row_values(rownum)[1] != "date":
            pay_data = xlrd.xldate_as_tuple(int(table.row_values(rownum)[1]), 0)
            if len(str(pay_data[1])) == 1:
                pay_data_row = str(pay_data[0])+"-0"+ str(pay_data[1])+"-"+ str(pay_data[2])
            else:
                pay_data_row = str(pay_data[0]) + "-" + str(pay_data[1]) + "-" + str(pay_data[2])

            if pay_data_row in activitiesDate:
                rmb = int(table.row_values(rownum)[2])
                pay_user_mumber = int(table.row_values(rownum)[3])
                pay_percent = float(pay_user_mumber) / summaryInfo[pay_data_row]['DailyEnter']

                arpu = float(rmb) / summaryInfo[pay_data_row]['DailyEnter']
                arppu = float(rmb) / pay_user_mumber
                summaryInfo[pay_data_row].setdefault("PayRMB", rmb)
                summaryInfo[pay_data_row].setdefault("PayUserNum",pay_user_mumber)
                summaryInfo[pay_data_row].setdefault("PayPercent", pay_percent)
                summaryInfo[pay_data_row].setdefault("ARPU", arpu)
                summaryInfo[pay_data_row].setdefault("ARPPU", arppu)

    fpdData = {}
    for rownum in range(tableFpd.nrows):
        if tableFpd.row_values(rownum)[4] != "DayNewPayingUserCount":
            pay_data = xlrd.xldate_as_tuple(int(tableFpd.row_values(rownum)[3]), 0)
            if len(str(pay_data[1])) == 1:
                pay_data_row = str(pay_data[0]) + "-0" + str(pay_data[1]) + "-" + str(pay_data[2])
            else:
                pay_data_row = str(pay_data[0]) + "-" + str(pay_data[1]) + "-" + str(pay_data[2])
            daily_newPay = 0
            if pay_data_row in activitiesDate:
                if not fpdData.has_key(pay_data_row):
                    fpdData.setdefault(pay_data_row,[int(tableFpd.row_values(rownum)[4])])
                else:
                    fpdData[pay_data_row].append(int(tableFpd.row_values(rownum)[4]))
    #写入总表
    for key in fpdData:
        summaryInfo[key].setdefault("DailyNewPay",sum(fpdData[key]))

    sortSummaryInfo = summaryInfo.items()
    sortSummaryInfo.sort()

    workbook = xlwt.Workbook(encoding='ascii')
    worksheet = workbook.add_sheet('summaryInfo')

    for i in range(len(title)):
        worksheet.write(0, i, title[i])

    count = 0
    for key in sortSummaryInfo:
        count += 1
        worksheet.write(count, 0, key[0])
        for subKey in key[1]:
            subCount = title.index(subKey)
            worksheet.write(count, subCount, key[1][subKey])

    workbook.save(activities_dir_name+"pay/SummaryInfo.xls")
    print "ok"

def ChangeDate(date):
    if len(str(date[1])) == 1 and len(str(date[2])) == 1:
        pay_data_row = str(date[0]) + "-0" + str(date[1]) + "-0" + str(date[2])
    elif len(str(date[1])) == 1:
        pay_data_row = str(date[0]) + "-0" + str(date[1]) + "-" + str(date[2])
    elif len(str(date[2])) == 1:
        pay_data_row = str(date[0]) + "-" + str(date[1]) + "-0" + str(date[2])
    else:
        pay_data_row = str(date[0]) + "-" + str(date[1]) + "-" + str(date[2])
    return pay_data_row



if __name__ == '__main__':
    print sys.argv
    if len(sys.argv) < 2:
        print 'No action specified.'
        sys.exit()
    if sys.argv[1].startswith('-'):
        option = sys.argv[1][1:]
        # fetch sys.argv[1] but without the first two characters
        if option == 'version':
            print 'Version 1.0'
        elif option == 'help':
            print '''
        This program will make Summary data.
        Options include:
          -version : Prints the version number
          -help    : Display this help
          -f       : Please Give me YinZehong's bigdata path '''
        elif option == 'f':
            if len(sys.argv) < 3:
                print "Give me right two path , channel path,YinZehong's bigdata path"
                sys.exit()
            #activities_dir_name = "/Users/tq/data_analysis/data_analysis_project/bigdata/"
            Summary_data(sys.argv[2]+"/")
        else:
            print 'Unknown option.'
        sys.exit()


