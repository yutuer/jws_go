###
// 对Date的扩展，将 Date 转化为指定格式的String
// 月(M)、日(d)、小时(h)、分(m)、秒(s)、季度(q) 可以用 1-2 个占位符，
// 年(y)可以用 1-4 个占位符，毫秒(S)只能用 1 个占位符(是 1-3 位的数字)
// 例子：
// (new Date()).Format("yyyy-MM-dd hh:mm:ss.S") ==> 2006-07-02 08:09:04.423
// (new Date()).Format("yyyy-M-d h:m:s.S")      ==> 2006-7-2 8:9:4.18
###
DateFormat = (date, fmt) ->
    if not date?
        date = new Date()

    oss = {
        "M+" : date.getMonth()+1,                 #月份
        "d+" : date.getDate(),                    #日
        "h+" : date.getHours(),                   #小时
        "H+" : date.getHours(),                   #小时
        "m+" : date.getMinutes(),                 #分
        "s+" : date.getSeconds(),                 #秒
        "q+" : Math.floor((date.getMonth()+3)/3), #季度
        "S"  : date.getMilliseconds()             #毫秒
    }
    if /(y+)/.test(fmt)
        fmt = fmt.replace(RegExp.$1, (date.getFullYear()+"").substr(4 - RegExp.$1.length))
    for k, v of oss
        if(new RegExp("(" + k + ")").test(fmt))
            fmt = fmt.replace(RegExp.$1, if RegExp.$1.length==1 then v else ("00"+ v).substr((""+ v).length))

    return fmt

TimeUnixToStr = (t) ->
    if t <= 0
        return ""
    d = new Date(t * 1000)
    res = d.getFullYear() + '-' +(d.getMonth() + 1)+'-'+d.getDate()+
          ' '+d.getHours() + ':'+d.getMinutes() + ':'+ d.getSeconds()
    return res


module.exports = {
    DateFormat    : DateFormat
    TimeUnixToStr : TimeUnixToStr
}
