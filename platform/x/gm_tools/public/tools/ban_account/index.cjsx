antd       = require 'antd'
React      = require 'react'
AccountKeyInput = require '../../common/account_input'
Api         = require '../api/api_ajax'

App = React.createClass
    getInitialState: () ->
        return {
            select_server : @props.curr_server
            player_to_send : ""
        }

    isServerRight : () -> @state.select_server? and @state.select_server.serverName?
    isAccountRight : () -> @refs.accountin? and @refs.accountin.IsRight()

    getLoadingState : () ->
        if not @isServerRight() or not @isAccountRight()
            return "disabled"
        return ''

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

    getServerName : () ->
        if @state.select_server?
            return @state.select_server.name
        else
            return ""

    ban : () ->
        api = new Api()
        api.Typ("banAccount")
           .ServerID(@state.select_server.serverName)
           .AccountID(@state.player_to_send)
           .Key(@props.curr_key)
           .Params(31536000)
           .Do (result) =>
                console.log "onSend"
                console.log result
                @props.onSend()

    ban1 : () ->
        api = new Api()
        api.Typ("banAccount")
           .ServerID(@state.select_server.serverName)
           .AccountID(@state.player_to_send)
           .Key(@props.curr_key)
           .Params(1)
           .Do (result) =>
                console.log "onSend"
                console.log result
                @props.onSend()                

    render:() ->
        <div>
            <AccountKeyInput {...@props} ref = "accountin" can_cb = {@handleUserChange}/>
            <antd.Button className={'ant-btn anticon-search ant-btn-primary ' + @getLoadingState()} onClick={@ban}>
                封禁账号1年
            </antd.Button>
            <antd.Button className={'ant-btn anticon-search ant-btn-primary ' + @getLoadingState()} onClick={@ban1}>
                解禁账号
            </antd.Button>
        </div>

module.exports = App
