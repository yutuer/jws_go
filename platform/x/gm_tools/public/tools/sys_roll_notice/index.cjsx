antd = require 'antd'
Api     = require '../api/api_ajax'
React      = require 'react'
AccountKeyInput = require '../../common/account_input'
State           = require '../../common/state'
ReactLoadingState = State.ReactLoadingState
NoticeTable     = require './notice_table'
NewSendModal    = require './new_sys_roll_notice'
Checkbox        = antd.Checkbox
CheckboxGroup   = Checkbox.Group

App = React.createClass
    getInitialState: () ->
        return {
            select_server : @props.curr_server
            player_to_send : ""
            vershard: []
            selected_sid    :[]
            loads : new ReactLoadingState(@)
        }

    LoadNotices: () ->
        if @refs.notice_table?
            @state.loads.EnterLoading("load")
            @refs.notice_table.Refersh 1, () =>
                @state.loads.LeaveLoading("load")

    isServerRight : () -> @state.selected_sid.length > 0

    getBntState : () ->
        return "ant-btn ant-btn-primary disabled" if not @isServerRight()
        return "ant-btn ant-btn-primary " + @state.loads.GetBtnStat("load")

    getTable : () ->
        if @state.loadstat is 0
            return <div>点击按钮加载ID</div>

        if not @isServerRight()
            return <div> </div>

        return <NoticeTable
                    {...@props}
                    ref        = "notice_table"
                    server_id  = @state.selected_sid
                    account_id = { @state.player_to_send } />

    getSendModal : () ->
        if not @isServerRight()
            return <div>请选择服务器</div>

        return <NewSendModal.NewModalWithButton
                    {...@props}
                    ref        = "notice_send"
                    modal_name = "添加跑马灯公告"
                    server_id  = @state.selected_sid
                    account_id = "ac"
                    on_loaded  = { @LoadNotices } />



    handleUserChange: (data) ->
        if data is ""
            data = "请输入玩家Id"
        @setState {
            player_to_send : data
        }


    selectSid : (e) ->
        ss = []
        for value in e
            ss.push(value)
        console.log "selectSid"
        console.log ss
        @setState {
            selected_sid:ss
        }

    genGroupCheckBox : (ar) ->
        plainOptions = []
        for a in ar
            plainOptions.push(a)
        return <tr>
                    <td>
                        <CheckboxGroup options={plainOptions} value={@state.selected_sid} onChange={@selectSid} />
                    </td>
               </tr>

    gencheckbox : () ->
        column_count = 5
        res = []
        selected = []
        count = 0
 
        res.push(@genGroupCheckBox(@state.vershard))
        return res

    selectallcheckbox : (e) ->
        if e.target.checked > 0
            @selectSid(@state.vershard)
        else
            @setState {
                selected_sid:[]
            }
        


    handleSubmit_ver : () ->
            api = new Api()
            api.Typ("getAllServer")
               .ServerID("")
               .AccountID("")
               .Key(@props.curr_key)
               .ParamArray([])
               .Do (result) =>
                    res = JSON.parse(result)
                    @setState {
                       vershard: res["Infos"]
                    }

    render : () ->
        <div>
            <div className="row-flex row-flex-middle row-flex-start" >
            <antd.Button
            onClick   = { @handleSubmit_ver } >
            获取所有服务器
            </antd.Button>
            </div>
            <label>
                <Checkbox onChange={@selectallcheckbox} />
                全选
            </label>
            <div  style={{ marginBottom: 24 }}>
                        <br />
                        </div>

                        <div>
                            {@gencheckbox()}
                        </div>


                        <div style={{ marginBottom: 24 }}>
                        <br />
                        </div>

            <div className="row-flex row-flex-middle row-flex-start" >
                <antd.Button
                     className = { @getBntState() }
                     onClick   = { @LoadNotices } >
                     加载跑马灯公告
                 </antd.Button>
                {@getSendModal()}
            </div>
            <hr />
            {@getTable()}
        </div>

module.exports = App
