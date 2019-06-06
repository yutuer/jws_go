# -*- coding: utf-8 -*-
import xlwt
import xlrd
import os
import sys


def Keep_data(channelPath,activities_dir_name):
    #渠道对照Map channelID --> channelName
    ChannelMap = {}
    #汇总后的总数据
    ChannelKeepDate = {}
    keepTitle = ["ChannelId","ChannelName","date","NewUser","OneKeep","TwoKeep","ThreeKeep","FourKeep","FiveKeep","SixKeep","SevenKeep"]
    indexKeep = [2, 4, 7, 10, 13, 16, 19]
    channel_data = xlrd.open_workbook(channelPath)
    #channel_data = xlrd.open_workbook("/Users/tq/data_analysis/data_analysis_project/channel.xlsx")
    tableChannel = channel_data.sheet_by_name(u'Sheet1')



    #读取渠道对照表
    for rownum in range(tableChannel.nrows):
        if tableChannel.row_values(rownum)[0] != "ServerChannel":
            ChannelMap.setdefault(int(tableChannel.row_values(rownum)[0]),tableChannel.row_values(rownum)[2])

    print "Channel Read ok"
    #读取留存原始数据
    rangeDir = os.listdir(activities_dir_name)
    for x in rangeDir:
        if x == "pay":
            continue
        keep_data = xlrd.open_workbook(activities_dir_name + "/" + x + "/retention_dayn.xlsx")
        table_keep = keep_data.sheet_by_name(u'retention_dayn_ch')

        keep_nrows = table_keep.nrows
        keep_ncols = table_keep.ncols
        #一个服的渠道很单元格
        tmp_channel = []
        for tmp_merged in table_keep.merged_cells:
            if tmp_merged[0] != 0:
                tmp_channel.append(tmp_merged)

        #每一个渠道的数据汇总
        for id in tmp_channel:
            if int(table_keep.cell_value(id[0],id[2])) == 0:
                continue
            key = int(table_keep.cell_value(id[0],id[2]))
            date_map = {}
            for num in range(id[0],id[1]):
                data = []
                num_date = xlrd.xldate_as_tuple(table_keep.row_values(num)[1], 0)

                pay_data_row = ChangeDate(num_date)

                for dex in indexKeep:
                    data.append(table_keep.row_values(num)[dex])
                date_map.setdefault(pay_data_row,data)
            ChannelKeepDate.setdefault(key,date_map)

    #写入数据
    workbook = xlwt.Workbook(encoding='ascii')
    worksheet = workbook.add_sheet('KeepInfo')

    for i in range(len(keepTitle)):
        worksheet.write(0, i, keepTitle[i])

    temp1 = 1
    temp2 = 1

    for key in ChannelKeepDate:
        key_len = len(ChannelKeepDate[key])
        if temp1 == temp2 :
            temp2 = key_len
        else:
            temp1 = temp2 + 1
            temp2 = temp1 + key_len

        worksheet.write_merge(temp1,temp2, 0, 0,key)
        worksheet.write_merge(temp1,temp2, 1, 1,ChannelMap[key])
        temp3 = temp1
        for subkey in ChannelKeepDate[key]:
            temp3 += 1
            worksheet.write(temp3, 2, subkey)
            for dex in range(len(ChannelKeepDate[key][subkey])):
                worksheet.write(temp3, dex + 3, ChannelKeepDate[key][subkey][dex])

    workbook.save(activities_dir_name+"/"+"pay/KeepData.xls")
    print "keep ok"


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
        This program will make keep data and need a channel.xlsx.
        Options include:
          -version : Prints the version number
          -help    : Display this help
          -f       : Please Give me channel path,YinZehong's bigdata path '''
        elif option == 'f':
            if len(sys.argv) < 4:
                print "Give me right two path , channel path,YinZehong's bigdata path"
                sys.exit()
            Keep_data(sys.argv[2],sys.argv[3])
        else:
            print 'Unknown option.'
        sys.exit()
