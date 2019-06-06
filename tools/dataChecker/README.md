# DataChecker

### 说明
用来验证数据正确性的工具

* config.toml为配置文件
* checklist.xlsx为执行的测试用例

### 命令行方式

`go run main.go -mode=checklist -checklist=targetlist`
- 执行targetlist.xlsx里的内容检查
- checklist可选

`go run main.go -mode=loc -language=en`
- 将活动表格内容转换为指定语言
- language选项为en, zh-Hans, ko等

`go run main.go -mode=loot`
- 生成目前所有关卡的掉落，目前设置为每个10万次

`go run main.go -mode=gacha`
- 生成目前所有gacha的掉落结果，目前设置为每个10万次
- 打印限时神将的暗控命中

`go run main.go -mode=hotactivity"
- 检查活动表中的内容是否有分组冲突和时间冲突


### 配置说明
见config.toml文件注释

### Checklist说明
见Checklist.xlsx第一页

### 检查类型说明

* 0-数值存在：检查文件B的数值，是否全部在文件A中有对应存在 - _避免id填写错误_
* 1-查重: 检查列里是否有重复的数值 - _避免数据重复时，后面的数据覆盖前面数据_
* 2-数值范围，确保数值正确
* 3-文件存在：检查列A里引用的文件，是否在工程目录中实际存在 - _避免资源文件名填写错误_
* Todo - 4-查空：检查列中是否有空值，或只有\t\n等白空格；以及IDS表格中，填写内容为IDS，或只有\t\n等白空格
* Todo - 5-时间：检查字符串格式的时间是否能转化为正确的时间 - _避免格式错误，或者生成 2017-2-30, 24:01:00, 12:60:00 之类的非法时间_
* 6-IDS存在：检查表格内引用的IDS，是否在对应的文本资源中存在

### 结果报告
如果有错误项的出现，会在log目录下生产md格式的验证报告，包括以下内容：
* 三个相关项目当前的git status
* 测试日期及时间
* 出现异常的checklist行号，对应的错误文件名、错误内容



