# -*- coding: utf-8 -*-
"""
功能：将Excel数据导入到MySQL数据库
"""
import xlrd
import MySQLdb
# Open the workbook and define the worksheet
book = xlrd.open_workbook("/Users/tq/Documents/serverStartTime.xlsx")
sheet = book.sheet_by_name("Sheet1")

#建立一个MySQL连接
database = MySQLdb.connect(host="54.223.192.252", user="test1", passwd="QmPhaQ8hYsxx", db="test")

# 获得游标对象, 用于逐行遍历数据库数据
cursor = database.cursor()

# 创建插入SQL语句
query = """INSERT INTO server_startTime (sid, start_time)
        VALUES (%s, %s)"""

# 创建一个for循环迭代读取xls文件每行数据的, 从第二行开始是要跳过标题
for r in range(1, sheet.nrows):
      product = sheet.cell(r,0).value
      customer = sheet.cell(r,1).value


      values = (product, customer)

      # 执行sql语句
      cursor.execute(query, values)

# 关闭游标
cursor.close()

# 提交
database.commit()

# 关闭数据库连接
database.close()

# 打印结果
print ""
print "Done! "
