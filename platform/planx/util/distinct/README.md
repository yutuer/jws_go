#distinct工具
###什么是distinct工具？
    distinct工具是一个专用于切片去重的工具，
    对外开放基于map[interface{}]bool实现的Set集合及
    ValuesAndDisinct(reply interface{}) ([]interface{}, error)函数
####set集合
    Set集合满足集合的基本特性（无序、唯一）
    Set集合的每一步写操作都使用了互斥锁，因此它是线程安全的。
####ValuesAndDisinct(reply interface{}) ([]interface{}, error)函数
    该函数将校验参数reply的类型，
    目前仅支持uint8\16\32\64|int8\16\32\64|string切片
    但允许玩家自己将参数先转换为[]interface{}再调用此函数
    如果此reply无法转换为[]interface{}则会产生一个error:
    "unexpected input,need slice"
    
    如果成功转换，该函数将直接去除reply中的重复元素，并返回去重后的结果和一个nil(error)