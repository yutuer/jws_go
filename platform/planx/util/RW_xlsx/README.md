#RW_xlsx工具
###什么是RW_xlsx工具?
    RW_xlsx即read/write_xlsxFile(读写xlsx文件)
    这是用来读取或写一个xlsx文件的工具包
###RW_xlsx工具提供了什么？
####RW_xlsx内存放了两个.go文件：
#####write_xlsx.go
    在write_xlsx.go只提供了一个函数：
    WriteToXlsxFile(data [][]string, XLSX_Name, sheet_name string)
    其中，data为用户希望写入的文件信息；
    XLSX_Name为用户希望写入的xlsx文件名称，
    sheet_name为用户希望写入的xlsx页签名称。
#####使用事例：
    WriteToXlsxFile(gainner_data, "gainer_"+
        time.Now().Format("02-01-2006_15:04:05_PM")+".xlsx", "gainner_orders")
    这是写一个名为gainer_02-01-2006_15:04:05_PM.xlsx的文件，其内容为gainner_data
#####read_xlsx.go
    在read_xlsx.go中提供了三个函数：
    
    GetFieldDataFromXlsx(XLSX_LOC, sheet, field string) ([]string, error)
    该函数传入xlsx文件名及页签名称和字段名称，获取该页签上指定字段的值
    返回两个参数，分别为xlsx文件sheet页签field字段中的数据[]string和错误信息
    
    
    GetDataFromXlsx(XLSX_LOC, sheet string) ([][]string, error)
    该函数传入xlsx文件名及页签名称，获取页面信息并返回
    返回两个参数，分别为xlsx文件sheet页签中的数据[][]string,
    如果有报错，返回error，其余返回值为零值
    
    
    GetIndexFromField(data [][]string, field string) (int, error)
    该函数根据传入的字段名查询该字段在信息中的索引
    找到则返回该索引且error为nil
    未查找到则返回-1,且error不为nil
#####使用事例：
    GetFieldDataFromXlsx("KoSuccessOrders.xlsx", "orders", "pay_description")
    该函数从KoSuccessOrders.xlsx的orders读取全部pay_description字段信息
    
    xlsx_data,_ := GetDataFromXlsx("KoSuccessOrders.xlsx", "orders")
    该函数从KoSuccessOrders.xlsx的orders读取全部信息并将结果存在xlsx_data中
    注意：忽略错误信息
    
    index, err := GetIndexFromField(xlsx_data, "pay_description")
    该函数从xlsx_data中获得pay_description的字段索引
    xlsx_data[i][pay_description]为pay_description字段的第i条信息。
    注意：i在pay_description之前。
    
    
