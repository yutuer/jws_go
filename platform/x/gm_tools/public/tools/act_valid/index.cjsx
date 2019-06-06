antd    = require 'antd'
Api     = require '../api/api_ajax'
React   = require 'react'
ReactBootstrap = require 'react-bootstrap'

Button          = ReactBootstrap.Button
Table           = antd.Table
Row             = antd.Row
Col             = antd.Col
Switch          = antd.Switch

App = React.createClass

    getInitialState: () ->
        _act_valid_id = {}
        _act_valid_id["绑定手机号"] = 0
        _act_valid_id["砸金蛋"] = 1
        _act_valid_id["运营活动-累计充值天数"] = 2
        _act_valid_id["运营活动-累计登录天数"] = 3
        _act_valid_id["运营活动-累计充值金额"] = 4
        _act_valid_id["运营活动-累计消费金额"] = 5
        _act_valid_id["运营活动-累计参与玩法次数"] = 6
        _act_valid_id["运营活动-累计购买资源次数"] = 7
        _act_valid_id["运营活动-每日累计充值金额"] = 8
        _act_valid_id["运营活动-每日累计消费金额"] = 9
        _act_valid_id["运营活动-将星之路"] = 10
        _act_valid_id["运营活动-限时神将"] = 11
        _act_valid_id["运营活动-招财猫"] = 12
        _act_valid_id["运营活动-攻城战"] = 13
        _act_valid_id["运营活动-限时商店"] = 14
        _act_valid_id["运营活动-投资名将"] = 15
        _act_valid_id["运营活动-单笔充值"] = 16
        _act_valid_id["运营活动-军团红包"] = 17
        _act_valid_id["运营活动-节日BOSS"] = 18
        _act_valid_id["运营活动-白盒宝箱"] = 19
        _act_valid_id["运营活动-开服七日红包"] = 20
        _act_valid_id["运营活动-开服七日排行榜"] = 21
        _act_valid_id["运营活动-运营排行榜"] = 22
        _act_valid_id["运营活动-(港澳台)FaceBook分享"] = 23
        _act_valid_id["运营活动-(港澳台)FaceBook邀请"] = 24
        _act_valid_id["运营活动-(港澳台)FaceBook关注"] = 25
        _act_valid_id["运营活动-(港澳台)跳转商店评分"] = 26
        _act_valid_id["运营活动-兑换商店"] = 27
        _act_valid_id["运营活动-神兵黑盒宝箱"] = 28
        _act_valid_id["运营活动-主将黑盒宝箱"] = 29
        _act_valid_id["运营活动-开服第二周排行榜"] = 30
        _act_valid_id["运营活动-EG账号绑定"] = 31
        _act_valid_id["运营活动-直购礼包"] = 32

        return {
            act_valid_id : _act_valid_id
            select_server : @props.curr_server
            onoff_info : [0]
        }

    handleServerChange: (data) ->
        console.log data
        @setState {
            select_server : data
        }

    load_data: () ->
        api = new Api()
        api.Typ("getActValidInfo")
           .ServerID(@state.select_server.serverName)
           .AccountID()
           .Key(@props.curr_key)
           .ParamArray()
           .Do (result) =>
                console.log result
                _info = result.split(",")
                info = []
                for i in _info
                    info.push(Number(i))
                console.log "load_data"+info
                @setState {
                    onoff_info:info
                }

    commit_data: () ->
        res = []
        for i in @state.onoff_info
            res.push(String(i))
        s = res.join(",")
        console.log s

        api = new Api()
        api.Typ("setActValidInfo")
           .ServerID(@state.select_server.serverName)
           .AccountID()
           .Key(@props.curr_key)
           .ParamArray([s])
           .Do (result) =>
                console.log result

    table_act_columns: () ->
        that = this
        return [{
          title: '活动名称',
          dataIndex: 'act'
        },{
          title: '开关状态',
          dataIndex: 'onoff'
          render : (text, record) ->
            s = false
            if that.state.onoff_info[Number(record.key)] > 0
                s = true 
            return <Switch defaultChecked={s} checked={s} onChange={
                (checked) =>
                    info = that.state.onoff_info
                    i = Number(record.key)
                    if checked
                        info[i] = 1
                    else
                        info[i] = 0
                    that.setState {
                        onoff_info:info,
                    }   
                } />
        }]

    table_act_content: () ->
        res = []
        console.log "table_act_content"+@state.act_valid_id
        for id_n, i of @state.act_valid_id
            console.log "table_act_content" +String(i)+" " +id_n
            res.push({
                key:i,
                act:id_n,
                })
        return res

    render:() ->
        <div>
            <p>用于设置活动开关</p>
            <br />
            
            <Row>
              <Col span="8" offset="2" >
                <Button bsStyle='primary' onClick= @load_data  >
                   加载指定服务器信息
                </Button>
              </Col>
              <Col span="4" offset="8">
                <Button bsStyle='primary' onClick= @commit_data  >
                   提交
                </Button>
              </Col>
            </Row>
            <br />

            <Table columns={@table_act_columns()} dataSource={@table_act_content()} 
                    pagination={{
                            showSizeChanger: true,
                            pageSize:20
                        }}/>
        </div>

module.exports = App