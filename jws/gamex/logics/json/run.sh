#!/usr/bin/env bash
#ClientPath="/Users/chiyutian/Client/trunk/Assets/Scripts/Network/NetLayer/Gen/gen"
#自己将这个变量设置成系统环境变量
#vim ~/.bash_profile
#重启gogland
./genserver -f
./genclient
echo $GOPATH
echo clientPath=$CLIENTPATH
mv -f *.cs $CLIENTPATH
rm -f *.cs