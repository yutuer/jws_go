CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o botx
#CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o botx
COPYFILE_DISABLE=1 tar cvzf botx.tar.gz botx replay resource conf
