dump.tar.gz中的scheme.json没有目录中的完整

但是里面主要是数据比较完整

# 批量创建数据库

```bash
python dynamodump.py -r cn-north-1 -m restore --schemaOnly -s Device_prod -d Devices_G201
python dynamodump.py -r cn-north-1 -m restore --schemaOnly -s Name_prod -d Name_G201
python dynamodump.py -r cn-north-1 -m restore --schemaOnly -s UserInfo_prod -d UserInfo_G201
```

```bash
python dynamodump.py -r cn-north-1 -m backup -s Mail_prod --schemaOnly
python dynamodump.py -r cn-north-1 -m backup -s UserShardInfo --schemaOnly
```



# Dump备份



ssh到一台能够访问读写这个数据库的机器

```bash
python dynamodump.py -r cn-north-1 -m backup -s Device_prod
python dynamodump.py -r cn-north-1 -m backup -s Name_prod
python dynamodump.py -r cn-north-1 -m backup -s UserInfo_prod
```

# Restore

Restore到需要充值返利的DynamoDB上的命令 并同时创建数据库

```bash
python dynamodump.py -r cn-north-1 -m restore -s Device_prod -d Devices_G202 --writeCapacity 100 --readCapacity 100 --log WARNING
python dynamodump.py -r cn-north-1 -m restore -s Name_prod -d Name_G202 --writeCapacity 100 --readCapacity 100 --log WARNING
python dynamodump.py -r cn-north-1 -m restore -s UserInfo_prod -d UserInfo_G202 --writeCapacity 100 --readCapacity 100 --log WARNING


python dynamodump.py -r cn-north-1 -m restore -s Device_prod -d Devices_G200 --writeCapacity 100 --readCapacity 100
python dynamodump.py -r cn-north-1 -m restore -s Name_prod -d Name_G200 --writeCapacity 100 --readCapacity 100
python dynamodump.py -r cn-north-1 -m restore -s UserInfo_prod -d UserInfo_G200 --writeCapacity 100 --readCapacity 100
```



## Restore with Dataonly

从Device_prod恢复到Devices_G201

```
python dynamodump.py --dataOnly -r cn-north-1 -m restore -s Device_prod -d Devices_G201 --log WARNING
```



然后导出Devices_G201验证导出数量

```
python dynamodump.py -r cn-north-1 -m backup -s Devices_G209 --dataOnly
cat dump/Devices_G209/data/*.json | jq '.Items|length'
```



