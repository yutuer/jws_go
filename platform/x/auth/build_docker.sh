ProjectPath=`echo $GOPATH | awk -F ":" '{print $1}'`
BuildImage="10.222.0.168:5000/golang:1.8.3-alpine"
BuildPath="vcs.taiyouxi.net/platform/x/auth"

docker run --rm -v $ProjectPath:/go -v $ProjectPath/src/vcs.taiyouxi.net/platform:/go/src/vcs.taiyouxi.net/platform -e GOPATH=/go $BuildImage go build -o /go/src/$BuildPath/auth_docker $BuildPath

tag=`date +%Y%m%d%H%M`
docker build --force-rm  --no-cache=true -t 10.222.0.168:5000/auth:0.0.0.$tag -f Dockerfile.auth .

rm $ProjectPath/src/$BuildPath/auth_docker
