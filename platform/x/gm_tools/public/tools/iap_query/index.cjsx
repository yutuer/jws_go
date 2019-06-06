antd    = require 'antd'
Api     = require '../api/api_ajax'
React   = require 'react'
ReactBootstrap = require 'react-bootstrap'
ProfileQuery = require '../../scripts/gm_profile_query'
AccountKeyInput = require '../../common/account_input'
TimeUtil     = require '../util/time'

GameServerList  = ProfileQuery.GameServerList
Input           = ReactBootstrap.Input
DatePicker      = antd.DatePicker
RangePicker     = DatePicker.RangePicker
FormItem        = antd.Form.Item
Row             = antd.Row
Col             = antd.Col
Form            = antd.Form
Button          = ReactBootstrap.Button
Table           = antd.Table

openNotificationWithIcon = (type, title, msg) ->
    notification[type] {
        message: title,
        description: msg
        duration: 5
    }

log_table_title = [{
    title:"充值时间",
    dataIndex: "paytime"
    },{
    title:"到账时间",
    dataIndex: "time"
    },{
    title:"角色ID",
    dataIndex: "acid"
    },{
    title:"账号名",
    dataIndex: "account_name"
    },{
    title:"角色名称",
    dataIndex: "name"
    },{
    title:"商品Id",
    dataIndex: "goodIdx"
        },{
    title:"商品name",
    dataIndex: "goodName"
        },{
    title:"充值金额",
    dataIndex: "money"
        },{
    title:"订单号",
    dataIndex: "order"
            },{
    title:"平台",
    dataIndex: "platform"
            },{
    title:"渠道",
    dataIndex: "channel"
            }]

mail_table_title = [{
    title:"mail时间",
    dataIndex: "mailtime"
    },{
    title:"充值时间",
    dataIndex: "paytime"
    },{
    title:"角色ID",
    dataIndex: "acid"
    },{
    title:"商品Id",
    dataIndex: "goodIdx"
        },{
    title:"充值金额",
    dataIndex: "money"
        },{
    title:"订单号",
    dataIndex: "order"
            },{
    title:"平台",
    dataIndex: "platform"
            },{
    title:"渠道",
    dataIndex: "channel"
            },{
    title:"是否领取",
    dataIndex: "isgot",
    filters: [{
        text: '已领取',
        value: 'true'
      }, {
        text: '未领取',
        value: 'false'
      }],
    onFilter : (value, record) ->
        return record.isgot.indexOf(value) == 0
}]

hc_table_title = [{
    title:"时间",
    dataIndex: "time"
    },{
    title:"角色ID",
    dataIndex: "acid"
    },{
    title:"获取/消耗",
    dataIndex: "gc",
    filters: [{
        text: '获取',
        value: 'Give'
      }, {
        text: '消耗',
        value: 'Cost'
      }],
    onFilter : (value, record) ->
        return record.gc.indexOf(value) == 0
    },{
    title:"hc类型",
    dataIndex: "hc_typ"
    },{
    title:"变化值",
    dataIndex: "chg"
    },{
    title:"之前值",
    dataIndex: "bef"
    },{
    title:"之后值",
    dataIndex: "aft"
    },{
    title:"变化途径",
    dataIndex: "reason"
    }]

gate_way_info = [{
    title:"账号Id",
    dataIndex: "acid"
    },{
    title:"订单号",
    dataIndex: "order"
    },{
    title:"流水号",
    dataIndex: "sn"
    },{
    title:"账号名",
    dataIndex: "ac_name"
    },{
    title:"角色名",
    dataIndex: "role_name"
    },{
    title:"商品Id",
    dataIndex: "goodIdx"
    },{
    title:"商品名",
    dataIndex: "good_name"
    },{
    title:"Rmb",
    dataIndex: "rmb"
    },{
    title:"平台",
    dataIndex: "platform"
    },{
    title:"渠道",
    dataIndex: "channel"
    },{
    title:"充值状态",
    dataIndex: "status"
    },{
    title:"通知状态",
    dataIndex: "tistatus"
    },{
    title:"HcBuy",
    dataIndex: "hcbuy"
    },{
    title:"HcGive",
    dataIndex: "hcgive"
    },{
    title:"支付时间",
    dataIndex: "paytime"
    },{
    title:"支付时间C",
    dataIndex: "paytime_c"
    },{
    title:"收到时间",
    dataIndex: "rectime"
    },{
    title:"is_test",
    dataIndex: "is_test",
    filters: [{
        text: '测试订单',
        value: 'Yes'
      }, {
        text: '正式订单',
        value: 'No'
      }],
    onFilter : (value, record) ->
        return record.is_test.indexOf(value) == 0
}]

App = React.createClass
    getInitialState: () ->
        {
          profile_key:""
          profile_valid:false
          start_time:null
          end_time:null
          query_log_res:[]
          query_mail_res:[]
          query_hc_res:[]
          query_gateway_res:[]
          select_server : @props.curr_server
          test_content:null
        }

    set_profile: (profile, is_can) ->
        if profile is ""
            profile = "请输入玩家Id"
        @setState {
          profile_key : profile
          profile_valid: is_can
        }

    is_can_query: () ->
        return @state.select_server? and @state.start_time? and @state.end_time? and @state.profile_valid

    is_can_query_not_finish: () ->
        return @state.select_server? and @state.profile_valid

    handleServerChange: (data) ->
        console.log data
        @setState {
            select_server : data
        }

    onChangeDate: (date, dateString) ->
        date[0].setHours(0)
        date[0].setMinutes(0)
        date[0].setSeconds(0)
        date[0].setMilliseconds(0)
        date[1].setHours(0)
        date[1].setMinutes(0)
        date[1].setSeconds(0)
        date[1].setMilliseconds(0)
        @setState {
            start_time: date[0]
            end_time: date[1]
        }

    query : () ->
        api = new Api()
        api.Typ("queryIAPInfo")
           .ServerID(@state.select_server.serverName)
           .AccountID(@state.profile_key)
           .Key(@props.curr_key)
           .ParamArray([@state.start_time.getFullYear()+"-"+(@state.start_time.getMonth() + 1)+"-"+@state.start_time.getDate(),
                 @state.end_time.getFullYear()+"-"+(@state.end_time.getMonth()+1)+"-"+@state.end_time.getDate()])
           .Do (result) =>
               console.log result
               res = JSON.parse(result)

               if res["ret"] != "ok"
                    openNotificationWithIcon("error", "Error", res["ret"])
               else
                    @setState {
                        query_log_res: res["queries_log"]
                    }

    query_not_finish : () ->
        api = new Api()
        api.Typ("queryIAPInfoMail")
           .ServerID(@state.select_server.serverName)
           .AccountID(@state.profile_key)
           .Key(@props.curr_key)
           .ParamArray()
           .Do (result) =>
               console.log result
               res = JSON.parse(result)
               if res["ret"] != "ok"
                    openNotificationWithIcon("error", "Error", res["ret"])
               else
                    @setState {
                        query_mail_res: res["queries_mail"]
                    }

    query_gateway : () ->
        sett = []
        if @state.start_time? and @state.end_time?
            sett = [String(@state.start_time.getTime()),String(@state.end_time.getTime())]
        api = new Api()
        api.Typ("queryGateWayInfo")
           .ServerID(@state.select_server.serverName)
           .AccountID(@state.profile_key)
           .Key(@props.curr_key)
           .ParamArray(sett)
           .Do (result) =>
               console.log result
               res = JSON.parse(result)
               if res["ret"] != "ok"
                    openNotificationWithIcon("error", "Error", res["ret"])
               else
                    @setState {
                        query_gateway_res: res["queries_android_pay"]
                    }

    query_hc : () ->
        api = new Api()
        api.Typ("queryHCInfo")
           .ServerID(@state.select_server.serverName)
           .AccountID(@state.profile_key)
           .Key(@props.curr_key)
           .ParamArray([@state.start_time.getFullYear()+"-"+(@state.start_time.getMonth() + 1)+"-"+@state.start_time.getDate(),
                 @state.end_time.getFullYear()+"-"+(@state.end_time.getMonth()+1)+"-"+@state.end_time.getDate()])
           .Do (result) =>
                console.log result
                res = JSON.parse(result)

                if res["ret"] != "ok"
                    openNotificationWithIcon("error", "Error", res["ret"])
                else
                    @setState {
                        query_hc_res:res["queries_hc"]
                    }

    table_log_content: () ->
        res = []
        i = 0
        for query in @state.query_log_res
            res.push({
                key:i,
                paytime:TimeUtil.TimeUnixToStr(query["PayTime"]),
                time:TimeUtil.TimeUnixToStr(query["Time"]),
                acid:query["AccountId"],
                account_name:query["AccountName"],
                name:query["Name"],
                goodIdx:query["GoodIdx"],
                goodName:query["GoodName"],
                money:query["Money"],
                order:query["Order"],
                platform:query["Platform"],
                channel:query["ChannelId"]
                })
            i += 1
        return res

    table_mail_content: () ->
        res = []
        i = 0
        for query in @state.query_mail_res
            res.push({
                key:i,
                paytime:TimeUtil.TimeUnixToStr(query["PayTime"]),
                mailtime:TimeUtil.TimeUnixToStr(query["Time"]),
                acid:query["AccountId"],
                goodIdx:query["GoodIdx"],
                money:query["Money"],
                order:query["Order"],
                platform:query["Platform"],
                channel:query["ChannelId"],
                isgot:String(query["IsGot"])
                })
            i += 1
        return res

    table_gateway_content: () ->
        res = []
        i = 0
        for query in @state.query_gateway_res
            pay_time_s = ""
            if query["pay_time_s"] != ""
                pay_time_s = TimeUtil.TimeUnixToStr(Number(query["pay_time_s"]))
            res.push({
                key:i,
                acid:query["uid"],
                order:query["order_no"],
                sn:String(query["sn"]),
                ac_name:query["account_name"],
                role_name:query["role_name"],
                goodIdx:query["good_idx"],
                good_name:query["good_name"],
                rmb:query["money_amount"],
                platform:query["platform"],
                channel:query["channel"],
                status:query["status"],
                tistatus:query["tistatus"],
                hcbuy:String(query["hc_buy"]),
                hcgive:String(query["hc_give"]),
                paytime:query["pay_time"],
                paytime_c:pay_time_s,
                rectime:TimeUtil.TimeUnixToStr(query["receiveTimestamp"]),
                is_test:query["is_test"]
                })
            i += 1
        return res

    table_hc_content : () ->
        res = []
        i = 0
        for query in @state.query_hc_res
            res.push({
                key:i,
                time:TimeUtil.TimeUnixToStr(query["Time"]),
                acid:query["AccountId"],
                gc:query["TypCG"],
                hc_typ:query["CurTyp"],
                chg:query["Chg"],
                bef:query["BefV"],
                aft:query["AftV"],
                reason:query["Reason"],
                })
            i += 1
        return res

    render:() ->
        <div>
            <div>
                <AccountKeyInput {...@props} ref = "accountin" can_cb = {@set_profile}/>
            </div>

            <div>
                <Row>
                    <Col span="10" offset="1">
                        <h4> 未/已到账支付查询（mail） </h4>
                    </Col>
                    <Col span="6" offset="1">
                        <Button bsStyle='primary' onClick= @query_not_finish disabled={ !@is_can_query_not_finish() } >
                            查询Mail
                        </Button>
                    </Col>
                </Row>
                <Row>
                    <Col span="10" offset="1">
                        注：mail只能查询到android支付的未/已到账信息
                    </Col>
                </Row>
                <Table columns={mail_table_title}
                     dataSource={@table_mail_content()} />
            </div>
            <Form horizontal >
                <FormItem label="查询日期选择(对下面三个表有效)：">
                  <Col span="6">
                    <RangePicker name="startDate" onChange={@onChangeDate} />
                  </Col>
                </FormItem>
            </Form>
            <div>
                <Row>
                    <Col span="10" offset="1">
                        <h4> Dynamo信息查询 </h4>
                    </Col>
                    <Col span="6" offset="1">
                        <Button bsStyle='primary' onClick= @query_gateway disabled={ !@is_can_query_not_finish() } >
                            查询
                        </Button>
                    </Col>
                </Row>
                <Row>
                    <Col span="10" offset="1">
                        注：包括android和ios的到账信息，和gateway的android的回调信息
                    </Col>
                </Row>
                <Table columns={gate_way_info}
                     dataSource={@table_gateway_content()}
                     pagination={{
                        total:@state.query_gateway_res.length,
                        showSizeChanger: true,
                        pageSize:20
                        }} />

            </div>
            <div>
                <Row>
                    <Col span="10" offset="1">
                        <h4> 已到账支付查询（log） </h4>
                    </Col>
                    <Col span="6" offset="1">
                        <Button bsStyle='primary' onClick= @query disabled={ !@is_can_query() } >
                            查询log
                        </Button>
                    </Col>
                </Row>
                <Table columns={log_table_title} dataSource={@table_log_content()}
                        pagination={{
                            total:@state.query_log_res.length,
                            showSizeChanger: true,
                            pageSize:20
                        }}  />
            </div>
            <div>
                <Row>
                    <Col span="10" offset="1">
                        <h4> HC情况查询 </h4>
                    </Col>
                    <Col span="6" offset="1">
                        <Button bsStyle='primary' onClick= @query_hc disabled={ !@is_can_query() } >
                            查询HC获得/消耗
                        </Button>
                    </Col>
                </Row>
                <Table columns={hc_table_title}
                     dataSource={@table_hc_content()}
                     pagination={{
                        total:@state.query_hc_res.length,
                        showSizeChanger: true,
                        pageSize:20
                        }} />
            </div>
        </div>

module.exports = App
