antd    = require 'antd'
Api     = require '../api/api_ajax'
React   = require 'react'
ReactBootstrap = require 'react-bootstrap'

Row             = antd.Row
Col             = antd.Col
Form            = antd.Form
FormItem        = antd.Form.Item
Button          = ReactBootstrap.Button
Table           = antd.Table
Modal           = antd.Modal
AntdButton      = antd.Button
Tag             = antd.Tag
Select          = antd.Select
Option          = antd.Option
DatePicker      = antd.DatePicker

multiLang = require '../../common/multi_lang'
LanguagePicker = multiLang.LanguagePicker
MultiLangPicker = multiLang.MultiLangPicker
langMap = multiLang.langMap

App = React.createClass

    getInitialState: () ->
        ss_init = {}
        ss_init["0"] = "新服"
        ss_init["1"] = "流畅"
        ss_init["2"] = "火爆"
        ss_init["3"] = "拥挤"
        ss_init["4"] = "维护"
        ss_init["5"] = "体验"
        ss_init["6"] = "玩家不可见"
        {
            shard_show_state_const:ss_init
            query_shards_res:[]
            machine_online:{}
            gamex_online:{}
            modal_visible:false
            modal_title:""
            oper_gid:""
            oper_sid:""
            oper_state:""
            dn_value:""
            oper_ip:""
            oper_sn:""
            oper_showstate:"0"
            oper_sevstarttime:""
            oper_machine:""
            lang_dn    : ""
            language   : @getLangMap(""),
            cur_lang   : "zh-Hans",
            multi_lang : ""
        }

    getLangMap: (lang)->
            langMap = {
                  "zh-Hans":"zh-Hans"
                  "zh-SG":"zh-SG"
                  "zh-HMT":"zh-HMT"
                  "en":"en"
            }
            tempLang = {}
            if !@IsJsonString(lang)
                for key of langMap
                     tempLang[key] = {dn:""}
            else
                tempLang = JSON.parse(lang)
            return tempLang

    IsJsonString : (jsonStr) ->
        if jsonStr == "" || !jsonStr
            return false
        try
            JSON.parse jsonStr
            return true
        catch err
            return false

    modalOk : () ->
        if @dn_validationState() is 'success'
            api = new Api()
            api.Typ("setShard2Etcd")
               .ServerID("")
               .AccountID("")
               .Key(@props.curr_key)
               .ParamArray([@state.oper_gid, @state.oper_sid, @state.dn_value,
                    @state.oper_state, @state.oper_ip, @state.oper_sn, @state.oper_showstate,
                    @state.oper_sevstarttime, @state.multi_lang, JSON.stringify(@state.language)])
               .Do (result) =>
                   if result == "ok"
                        querys = @state.query_shards_res
                        for query in querys
                            if String(query["Gid"]) == @state.oper_gid and String(query["Sid"]) == @state.oper_sid
                                query["DName"] = @state.dn_value
                                query["State"] = @state.oper_state
                                query["ShowState"] = @state.oper_showstate
                                query["Language"] = JSON.stringify(@state.language)
                                query["MultiLang"] = @state.multi_lang
                                @setState {
                                    query_shards_res:querys
                                }
                                break

            @setState {
                modal_visible:false
            }
        
        
    modalCancel : () ->
        @setState {
            modal_visible:false
        }

    query_shards : () ->
        api = new Api()
        api.Typ("getShardsFromEtcd")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray()
           .Do (result) =>
               console.log result
               res = JSON.parse(result)
               machine_onl = {}
               gamex_onl = {}
               for e in res["Shards"]
                  if e["Machine"] != "" 
                    machine_onl[e["Sid"]] = e["Machine"]
                  if e["SName"] != "" and e["State"] == "online" and e["Ip"] != "" 
                    gamex_onl[e["Sid"]] = e["Ip"]
               @setState {
                   query_shards_res: res["Shards"]
                   machine_online: machine_onl
                   gamex_online:gamex_onl
               }

    table_shards_columns: () ->
        that = this
        return [{
          title: '机器IP',
          dataIndex: 'machine'
          sorter : (a, b) ->
            return a.machine.localeCompare(b.machine)
        },{
          title: 'Gid',
          dataIndex: 'gid',
          sorter : (a, b) ->
            return a.gid - b.gid
        }, {
          title: 'Sid',
          dataIndex: 'sid'
          sorter : (a, b) ->
            return a.sid.localeCompare(b.sid)
        }, {
          title: '游戏服显示名称',
          dataIndex: 'dn'
        }, {
          title: '服务器显示状态',
          dataIndex: 'ss'
          render : (text, record) ->
            if record.ss == "维护"
              return <div color="red"> {record.ss} </div>
            return <div> {record.ss} </div>
        },{
          title: '游戏服IP',
          dataIndex: 'ip'
          sorter : (a, b) ->
            return a.ip.localeCompare(b.ip)
        }, {
          title: 'ShardName',
          dataIndex: 'sn'
        }, {
          title: '设置的开服时间',
          dataIndex: 'starttime'
        },  {
          title: '启动时间',
          dataIndex: 'lauchtime'
        },  {
            title:'开服时间',
            dataIndex:'realstarttime'
        },  {
            title:'服务器版本号',
            dataIndex:'version'
        },  {
          title: '玩家可见',
          dataIndex: 'state'
        }, {
          title: '玩家可进',
          dataIndex: 'gs'
          render : (text, record) ->
            color = "red"
            title = "No"
            if that.state.gamex_online[record.sid]? and record.ss != "维护"
              color = "green"
              title = "Yes"
            return <Tag color={color}>{title}</Tag>
        },{
          title: '操作',
          key: 'operation',
          render : (text, record) ->
            if record.sn != ""
              return <Button bsStyle='primary' onClick= {
                      () =>
                          show_state_id = "0"
                          for id of that.state.shard_show_state_const
                            if that.state.shard_show_state_const[id] == record.ss
                              show_state_id = id
                              break
                          t = "大区:"+record.gid+"   服务器Id:"+record.sid
                          operstate = ""
                          if record.state == "Yes"
                            operstate = "online"
                          tempLangMap = that.getLangMap(record.language)
                          initLang = "zh-Hans"
                          defaultML = "0"
                          if record.multi_lang == "1"
                              defaultML = "1"
                          that.setState {
                              modal_title:t
                              oper_gid:record.gid
                              oper_sid:record.sid
                              oper_state:operstate
                              dn_value:record.dn
                              oper_ip:record.ip
                              oper_sn:record.sn
                              oper_showstate:show_state_id
                              oper_sevstarttime:record.starttime
                              oper_machine:record.machine
                              modal_visible:true
                              language: tempLangMap
                              lang_dn: tempLangMap[initLang].dn
                              cur_lang: initLang
                              multi_lang: defaultML
                          }
                      }  >
                        操作
                     </Button>
            else
              return <div> </div>
            
        }]

    table_shards_content: () ->
        res = []
        i = 0
       
        for query in @state.query_shards_res
            show_state = ""
            id = query["ShowState"]

            if id?
              f = @state.shard_show_state_const[id]
              if f?
                show_state = f

            st = ""
            if query["State"] == "online"
              st = "Yes"

            res.push({
                key:i,
                machine:query["Machine"],
                gid:String(query["Gid"]),
                sid:String(query["Sid"]),
                dn:query["DName"],
                ip:query["Ip"],
                sn:query["SName"],
                starttime:query["StartTime"],
                lauchtime:query["LaunchTime"],
                realstarttime:query["RealStartTime"],
                version:query["Version"],
                state:st,
                ss:show_state,
                language:query["Language"],
                multi_lang:query["MultiLang"],
                })
            i += 1
        return res

    checkDN: (rule, value, callback) ->
        if value.length <= 0
            callback("不能为空")

    dn_handleChange : (event) ->
        @setState {
          dn_value: event.target.value
        }

    dn_validationState : () ->
        v =  @state.dn_value
        if v.length > 0
            return 'success'
        else
            return 'error'

    onModalSubmit : () ->
        console.log "onModalSubmit"

        console.log @state
        api = new Api()
        api.Typ("setShardStateEtcd")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray([@state.oper_gid, (@state.oper_sid-1).toString(), "2"])
           .Do (result) =>
                console.log result
        @setState {
            oper_state:"online"
        }

    onSerRestart :() ->
        api = new Api()
        api.Typ("restartShard")
          .ServerID("")
          .AccountID("")
          .Key(@props.curr_key)
          .ParamArray([@state.oper_gid, @state.oper_sid, @state.oper_machine])
          .Do (result) =>
              console.log result

    select_op : () ->

        res = []
        for k,v of @state.shard_show_state_const
            res.push(<Option value={k}>{v}</Option>)

        return res

    onSelectChange : (value) ->
        console.log "on"
        console.log value
        @setState {
            oper_showstate: value
        }

    onChangeSevStartDate: (date) ->
        dateStr = date.getFullYear()+"/"+(date.getMonth()+1)+"/"+date.getDate()
        console.log dateStr
        @setState {
            oper_sevstarttime:dateStr
        }
    getInitDatePicker : () ->
        @state.oper_sevstarttime.slice(0, 10)

    handleMultiLangChange: (value) ->
        @setState {multi_lang    : value}

    handleLanguageChange: (value) ->
        tempMap = @state.language
        @setState {
            cur_lang  : value,
            lang_dn      : tempMap[value].dn,
        }

    handleLangDnChange:(event) ->
        tempMap = @state.language
        tempMap[@state.cur_lang].dn = event.target.value
        @setState { lang_dn       : event.target.value, language : tempMap}

    render:() ->
        <div>
            <p>说明：</p>
            <p>玩家可见：在游戏服务器列表中是否可以见到此服务器</p>
            <p>玩家可进：选中此服务器后，是否可以成功进入游戏</p>
            <p>当服务器没有上线过时，可以进行重启操作（此操作不要连续点），之后QA进行账号白名单测试，测试完成后，上线则玩家可进了</p>
            <p>[设置的开服时间]是运营设置的时间，[启动时间]是服务器启动时间，[开服时间]是服务器运营活动实际用的时间</p>
            <p>[设置的开服时间]应该等于[开服时间]，并早于[启动时间]才正确</p>
            <br />
            <Button bsStyle='primary' onClick= @query_shards  >
               获取Shards信息
            </Button>
            <Table columns={@table_shards_columns()} dataSource={@table_shards_content()} />
            <Modal title={@state.modal_title} visible={@state.modal_visible}
              onOk={@modalOk} onCancel={@modalCancel}>
              <div>
                <ReactBootstrap.Input
                    label="服务器显示名称（display name）:"
                    type='text'
                    value={@state.dn_value}
                    bsStyle={@dn_validationState()}
                    hasFeedback
                    ref='dn_input'
                    groupClassName='group-class'
                    labelClassName='label-class'
                    onChange={@dn_handleChange} />
               </div>

               <div className="row-flex row-flex-middle row-flex-start">
                   <div>多语言:</div>
                   <MultiLangPicker defaultValue = {@state.multi_lang} value = {@state.multi_lang} handleChange={@handleMultiLangChange}></MultiLangPicker>
                   <div>语言类型:</div>
                   <LanguagePicker defaultValue = {@state.cur_lang} handleChange={@handleLanguageChange}></LanguagePicker>
               </div>

               <div className="row-flex row-flex-middle row-flex-start">
                    <div>服务器名字:</div>
                    <input className="ant-input" value = {@state.lang_dn} onChange={@handleLangDnChange}></input>
               </div>

               <div>
                  <Row>
                    <Col span="6" offset="1">服务器显示状态：</Col>
                  </Row>
                </div>

                <div>
                  <Row>
                    <Col span="6" offset="1">
                      <Select defaultValue={@state.oper_showstate} style={{ width: 120 }} onSelect={@onSelectChange}>
                        {@select_op()}
                      </Select>
                    </Col>
                  </Row>
                </div>
                <p><br/></p>
                
                <Form horizontal >
                  <FormItem label="设置服务器开服时间：">
                    <Col span="6">
                      <DatePicker name="sevstarttime" value={@getInitDatePicker()} format="yyyy/MM/dd" onChange={@onChangeSevStartDate} />
                    </Col>
                  </FormItem>
                </Form>
                <div>
                  <AntdButton type="primary" 
                          disabled={@state.oper_state == "online" or @state.oper_ip == "" or @state.oper_sn == "" or @state.oper_machine == ""}
                          onClick=@onSerRestart >
                          服务器重启
                  </AntdButton>
                </div>
                <p><br/></p>
                <div>
                  <Row>
                    <Col span="5" offset="2">
                      <AntdButton type="primary"
                          disabled={@state.oper_state == "online" or @state.oper_ip == "" or @state.oper_sn == ""}
                          onClick=@onModalSubmit >
                          服务器上线
                      </AntdButton>
                    </Col>
                    <Col span="8">
                      {
                          if @state.oper_state == "online"
                              "服务器已上线"
                          if @state.oper_ip == "" or @state.oper_sn == ""
                              "服务器未正常启动, 不能上线"
                      }
                    </Col>
                  </Row>
                </div>
            </Modal>
        </div>

module.exports = App