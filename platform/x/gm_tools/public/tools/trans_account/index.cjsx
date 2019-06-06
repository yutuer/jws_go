antd         = require 'antd'
Api          = require '../api/api_ajax'
AccountKeyInput = require '../../common/account_input'
React      = require 'react'
ReactBootstrap = require 'react-bootstrap'
Upload = antd.Upload
Icon   = antd.Icon
AntButton = antd.Button

ButtonToolbar  = ReactBootstrap.ButtonToolbar
Button         = ReactBootstrap.Button
MenuItem       = ReactBootstrap.MenuItem
DropdownButton = ReactBootstrap.DropdownButton
Table          = ReactBootstrap.Table
ModalTrigger   = ReactBootstrap.ModalTrigger
Modal          = ReactBootstrap.Modal
Navbar         = ReactBootstrap.Navbar
Nav            = ReactBootstrap.Nav
NavItem        = ReactBootstrap.NavItem
Input          = ReactBootstrap.Input
ListGroup      = ReactBootstrap.ListGroup
ListGroupItem  = ReactBootstrap.ListGroupItem


App = React.createClass
    getInitialState: () ->
        return {
            select_server  : @props.curr_server
            player_to_send : ""
            account_data   : ""
        }

    isServerRight  : () -> @state.select_server? and @state.select_server.serverName?
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


    getServerName : () ->
        if @state.select_server?
            return @state.select_server.name
        else
            return ""

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
        @post "../api/v1/command/" + @state.select_server.serverName + "/" + @state.player_to_send  + "/" + "getAllInfoByAccountID",
            key: @props.curr_key
            params: @state.send

    mod : () ->
        api = new Api()
        console.log @state.send
        api.Typ("setAllInfoByAccountID")
           .ServerID(@state.select_server.serverName)
           .AccountID(@state.player_to_send)
           .Key(@props.curr_key)
           .Params(@state.account_data)
           .Do (result) =>
                console.log "on getAllInfo"
                console.log result

    cleanUnEquipItem : () ->
        api = new Api()
        console.log @state.send
        api.Typ("cleanUnEquipItems")
           .ServerID(@state.select_server.serverName)
           .AccountID(@state.player_to_send)
           .Key(@props.curr_key)
           .Params()
           .Do (result) =>
                console.log "on getAllInfo"
                console.log result


    handleChangeDataForCommit : (event) ->
        @setState {
            account_data : event.target.value
        }

    getDataForCommit : () -> @state.account_data
    getServerName : () ->
         if @state.select_server == undefined
            serverName = ""
         else
            serverName = @state.select_server.serverName
         return serverName
    render:() ->
        <div>
            <AccountKeyInput {...@props} ref = "accountin" can_cb = {@handleUserChange}/>
            <Button
                bsStyle='primary'
                onClick= {@query}
            >
                Query
            </Button>

            <Upload {...@props}


                   action = {"../api/v1/command/" + @getServerName() + "/" + @state.player_to_send  + "/" + "setAllInfoByAccountID"}
                   data = {"key": @props.curr_key, "params": []}
                   >
                   <Button type="ghost">
                       <Icon type="upload"/> 上传账号信息
                   </Button>
            </Upload>
            <hr />
            <Button
                bsStyle='danger'
                onClick= {@cleanUnEquipItem}
            >
                CleanUnEquipItem
            </Button>

            <input type="text" value={@getDataForCommit()} onChange={@handleChangeDataForCommit} />
        </div>

module.exports = App
