NewMailAllSendModal = require './new_mail_to_all'

maillist = require './mail_list'
MailList = maillist.MailList
BatchMailResult = maillist.BatchMailResult
React               = require 'react'
AccountKeyInput     = require '../../common/account_input'
boot                = require 'react-bootstrap'
antd       = require 'antd'
Upload = antd.Upload
Icon   = antd.Icon
Button = antd.Button
Input  = antd.Input
Api     = require '../api/api_ajax'

MailAll = React.createClass
    getInitialState: () ->
        return {
            select_server : @props.curr_server
            player_to_send : "0:0:1331"
        }

    handleServerChange: (data) ->
        console.log data
        @setState {
            select_server : data
        }
        setTimeout @OnSend, 100

    handleUserChange: (data) ->
        console.log data
        if data is ""
            data = "请输入玩家Id"
        @setState {
            player_to_send : data
        }

    OnSend: () ->
        console.log "SendOver"
        @refs.mail_list_all.Refersh(1)

    getAllMailName : () ->
        if @state.select_server?
            res = @state.select_server.serverName.split ':'
            return "all:" + res[1]
        else
            return "all:10"

    getServerName : () ->
        if @state.select_server?
            return @state.select_server.serverName
        else
            return ""

    render:() ->
        <div>
            <NewMailAllSendModal
                modal_name="添加全服邮件"
                mail_name={@getAllMailName()}
                server_name={@getServerName()}
                onSend={@OnSend}/>
            <hr />
            <div>全服邮件:</div>
            <MailList
                {...@props}
                ref="mail_list_all"
                server_name={@getServerName()}
                is_personal = false
                mail_name={@getAllMailName()}/>
            <hr />
        </div>


MailOne = React.createClass
    getInitialState: () ->
        return {
            select_server : @props.curr_server
            player_to_send : "0:0:1331"
        }

    handleServerChange: (data) ->
        console.log data
        @setState {
            select_server : data
        }
        setTimeout @OnSend, 100


    handleUserChange: (data) ->
        console.log data
        if data is ""
            data = "请输入玩家Id"
        @setState {
            player_to_send : data
        }

    OnSend: () ->
        console.log "SendOver"
        @refs.mail_list.Refersh(1)

    getAllMailName : () ->
        if @state.select_server?
            res = @state.select_server.serverName.split ':'
            return "all:" + res[1]
        else
            return "all:10"

    getServerName : () ->
        if @state.select_server?
            return @state.select_server.serverName
        else
            return ""

    render:() ->
        <div>
            <AccountKeyInput can_cb = {@handleUserChange}/>

            <div className = "row-flex row-flex-middle row-flex-start">
                <NewMailAllSendModal
                    modal_name={"添加单人邮件给:" + @state.player_to_send}
                    mail_name={"profile:"+@state.player_to_send}
                    server_name={@getServerName()}
                    onSend={@OnSend}/>
                <button className="ant-btn ant-btn-primary" onClick={@OnSend}>
                    查询个人所有邮件
                </button>
            </div>

            <hr/>

            <div>{"单人邮件[" + @state.player_to_send + "]:"}</div>
            <MailList ref="mail_list"
                {...@props}
                server_name={@getServerName()}
                is_personal = true
                mail_name={"profile:"+@state.player_to_send}/>
            <hr />
        </div>

MailBatch = React.createClass
    getInitialState: () ->
        api = new Api()
        api.Typ("getBatchMailIndex")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray([])
           .Do (result) =>
                @setState {
                   batch_index: "当前批量号: " + result
                }
        return {
            batch_query_index : ""
            batch_index : ""
        }

    onSuccess: (data) ->
        if data.file.status == "done"
            @refs.batch_result.Refresh(data.file.response)
            @setState {
                batch_index: "当前批量号: " + data.file.response.CurrentBatchIndex
            }
        if data.file.status == "error"
            alert data.file.response

    post : (url, params) ->
        temp = document.createElement('form')
        temp.action = url
        temp.method = 'post'
        temp.style.display = 'none'
        for x of params
            opt = document.createElement('input')
            opt.name = x
            opt.value = params[x]
            temp.appendChild opt
        document.body.appendChild temp
        temp.submit()
        temp

    query : () ->
        data = {
            "params" : @params,
            "key"    : @key
        }
        @post "../api/v1/command/null/null" + "/" + "getBatchMailDetail",
            key: @props.curr_key
            params: [@state.batch_query_index]

    onChange: (e) ->
        @setState {
            batch_query_index: e.target.value
        }

    render:() ->
        <div>
           <div>
                <Input value={@state.batch_index} disabled="disabled"/>
           </div>
            <hr/>
            <div>
                <Upload {...@props}
                       action = "/api/v1/batchmail"
                       onChange = {@onSuccess}>
                       <Button type="ghost">
                           <Icon type="upload"/> 批量发送邮件
                       </Button>
                </Upload>
            </div>
            <hr/>
            <div>
                <Upload {...@props}
                       action = "/api/v1/batch_del_mail"
                       onChange = {@onSuccess}>
                       <Button type="ghost">
                           <Icon type="upload"/> 批量删除邮件
                       </Button>
                </Upload>
            </div>
            <hr/>
            <div>
                <Input onChange={@onChange}/>
                <Button bsStyle='primary' onClick= {@query}>
                    下载批量邮件详情
                </Button>
            </div>
            <hr/>
            <div>
                <BatchMailResult ref="batch_result"/>
            </div>
        </div>

module.exports = {
    "MailAll" : MailAll
    "MailOne" : MailOne
    "MailBatch": MailBatch
}
