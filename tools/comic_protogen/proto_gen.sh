#!/usr/bin/env bash

go build
cp  comic_protogen ~/comicclient/tools/protogen/

#destRoot="./"
#rm -rf gen_server_temp gen_client_temp
#./comic_protogen -server -client -d ./proto

#cp -rn ./gen_server_temp/handlers ${destRoot}

#cp -f ./gen_server_temp/gen_func.go ./gen_server_temp/gen_reg_func.go ${destRoot}
