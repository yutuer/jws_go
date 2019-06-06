MONGOURL="mongodb://mongouser:XXXXX@10.30.4.202:27017/?authSource=admin"
./dynamo2mongodb dump/Device_prod/data/ "${MONGOURL}" AuthDB_G203
./dynamo2mongodb dump/Device_prod/data/ "${MONGOURL}" AuthDB_G2