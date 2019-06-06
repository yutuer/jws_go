export GOOS="linux"
rm -rf tmp_crossservice_linux
go build -o crossservice_linux_tmp
mkdir tmp_crossservice_linux
mv crossservice_linux_tmp ./tmp_crossservice_linux/crossservice_linux
mkdir tmp_crossservice_linux/conf
cp conf/*.toml tmp_crossservice_linux/conf/
cp conf/*.xml tmp_crossservice_linux/conf/
mkdir tmp_crossservice_linux/conf/data
cp ../gamex/conf/data/* tmp_crossservice_linux/conf/data/
tar czf crossservice_linux.tar.gz tmp_crossservice_linux
rm -rf tmp_crossservice_linux