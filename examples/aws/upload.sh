GOOS=linux GOARCH=amd64  go build dynamodb.go
scp dynamodb root@10.30.0.186:/opt/supervisor/
