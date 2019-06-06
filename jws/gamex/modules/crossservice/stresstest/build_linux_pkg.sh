export GOOS="linux"
go build -o stresstest_linux
tar czf stresstest_linux.tar.gz stresstest_linux config.json
# scp stresstest_linux.tar.gz ec2-user@54.223.61.58:/home/ec2-user
# scp stresstest_linux.tar.gz ec2-user@54.223.114.207:/home/ec2-user