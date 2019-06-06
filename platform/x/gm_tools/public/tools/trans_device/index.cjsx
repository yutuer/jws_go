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
            acid1 : ""
            acid2 : ""
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


    handleUserChange1: (data) ->
        console.log data
        if data is ""
            data = "请输入玩家Id"
        @setState {
            acid1 : data
        }

    handleUserChange2: (data) ->
        console.log data
        if data is ""
            data = "请输入玩家Id"
        @setState {
            acid2 : data
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
            acidArray = [@state.acid1, @state.acid2]
            api = new Api()
            console.log @state.send
            api.Typ("transDevice")
               .ServerID(@state.select_server.serverName)
               .AccountID("AID")
               .Key(@props.curr_key)
               .ParamArray(acidArray)
               .Do (result) =>
                    console.log "onSend"
                    console.log result
                    @setState {
                        res : JSON.parse(result)
                    }

    render:() ->
        <div>
            <AccountKeyInput {...@props} ref = "accountin" can_cb = {@handleUserChange1}/>
            <AccountKeyInput {...@props} ref = "accountin" can_cb = {@handleUserChange2}/>
            <Button bsStyle='primary' onClick= {@query}>替换</Button>
        </div>

module.exports = App
