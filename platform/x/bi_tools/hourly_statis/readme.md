给发行的BI数据统计，每小时统计一次，并生成一个文件里

分别从elasticsearch和graphite获取数据

debug:发 SIGHUP 信号

运行时 -t 参数，可以带个固定时间（2006-01-02 15），发信号生成此小时的数据