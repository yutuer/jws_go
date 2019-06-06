# dev 10.222.3.10 10->3
# qa 10.222.3.10 14->10

array=(
1:14:37519ced-0703-4103-b6d3-d1ce844ab299
#0:10:88f85ef0-62a7-4f19-9a23-c440869b9bc1
)
dbpwd=""
#导出
for i in "${array[@]}"
do
  echo $i
	echo "./tool_account2json 10.222.3.10:6379 10 $dbpwd $i > $i.json"
  ./tool_account2json 10.222.3.10:6379 10 $dbpwd $i > $i.json
done

# 手动修复

#导入数据库，修复

#for i in "${array[@]}"
#do
#  echo $i
#  echo "./json2account 10.222.3.10:6379 3 $dbpwd $i.json $i"
#  ./json2account 10.222.3.10:6379 3 $dbpwd $i.json $i
#done
