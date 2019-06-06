#!/usr/bin/env bash
db=("AuthDB_G1" "AuthDB_G1" "AuthDB_G1" "QuickPay_G2")
collection=("Devices" "UserInfos" "UserShardInfos" "Pay")
host="10.30.4.203:27017"
for i in "${!db[@]}"; do
    echo mongoexport db: ${db[$i]} collection: ${collection[$i]}
    /mongodb/mongodb-linux-x86_64-3.4.2/bin/mongoexport -u mongouser -p eC3XRJDhn6f3  --authenticationDatabase=admin\
    --db ${db[$i]} --collection ${collection[$i]} --out json/${collection[$i]}.json --jsonArray\
    --host $host
done
