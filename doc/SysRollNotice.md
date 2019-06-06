# 跑马灯

## 概要设计

- 广播消息全部从gamex服务器触发
- 使用城镇广播服务器进行消息广播，gamex和城镇广播服务器之间采用RestApi方式进行通信
- GM平台触发的跑马灯也是通过城镇广播服务器，采用RestApi方式进行通信；GM定时发的job由GM工具自己实现
- 各个跑马灯的触发条件，服务器写死，具体参数配在CommonConfig中

## 部署相关
- gamex服务器，配置gate.toml的broadCast_url，如：broadCast_url="http://{server:port}/broadcast"，其中server:port为城镇广播服务器的地址和端口
- 城镇广播服务器的url为“/broadcast”，采用http的post方式，httpbody为json，正常返回httpcode：200，并返回json的map，其中包含“status”的key对应的value若为“ok”；错误则错误码400，并“status”的key对应的value为错误信息
- 城镇广播服务器的url为“/broadcast”中的json格式为

```json
{

	type BroadCastMsg struct {
		Type string // 消息类型，如跑马灯就是 SysRollNotice 
		Msg  string // 消息体的json
	}
	
}
```


