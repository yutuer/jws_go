antd         = require 'antd'
Api          = require '../api/api_ajax'
AccountKeyInput = require '../../common/account_input'
React      = require 'react'
ReactBootstrap = require 'react-bootstrap'

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
Grid           = ReactBootstrap.Grid
Row            = ReactBootstrap.Row
Col            = ReactBootstrap.Col

NamePassEditor = require './name_editor'
DeviceEditor   = require './device_editor'

App = React.createClass
    getInitialState: () ->
        return {
            select_server  : @props.curr_server
            player_to_send : ""
            name_to_find   : ""
            device_to_find : ""
            account_data   : {}
        }

    handleServerChange: (data) ->

    handleNameChange: () ->
        @setState {
              name_to_find: @refs.name_input.getValue()
        }

    handleDeviceChange: () ->
        @setState {
              device_to_find: @refs.device_input.getValue()
        }

    getServerName : () ->
        if @state.select_server?
            return @state.select_server.name
        else
            return ""

    queryByDeciveID : (  ) ->
        id = @refs.device_input.getValue()
        api = new Api()
        console.log @state.send
        api.Typ("getAccountByDeviceID")
           .ServerID("SID")
           .AccountID("AID")
           .Key(@props.curr_key)
           .Params(id)
           .Do (result) =>
                console.log "on queryByDeciveID"
                console.log result
                @setState {
                    account_data : JSON.parse(result)
                }

    queryByName : ( ) ->
        name = @refs.name_input.getValue()
        api = new Api()
        console.log @state.send
        api.Typ("getAccountByName")
           .ServerID("SID")
           .AccountID("AID")
           .Key(@props.curr_key)
           .Params(name)
           .Do (result) =>
                console.log "on queryByName"
                console.log result
                @setState {
                    account_data : JSON.parse(result)
                }

    getAccountRes : () ->
        return @state.account_data.user_id if @state.account_data? and @state.account_data.user_id?
        return "null"


    render:() ->
        <div>
            <div>
                <Input
                    type='text'
                    value={@state.name_to_find}
                    placeholder='请输入玩家注册名...'
                    help='通过用户名查找UID, UID在下面显示'
                    bsStyle='success'
                    hasFeedback
                    ref='name_input'
                    onChange={@handleNameChange} />
                <Button
                    bsStyle='primary'
                    onClick= {@queryByName}
                >
                    查找注册名
                </Button>
            </div>
            <div>
                <Input
                    type='text'
                    value={@state.device_to_find}
                    placeholder='请输入DeviceID...'
                    help='通过DeviceID查找UID, UID在下面显示'
                    bsStyle='success'
                    hasFeedback
                    ref='device_input'
                    onChange={@handleDeviceChange} />
                <Button
                    bsStyle='primary'
                    onClick= {@queryByDeciveID}
                >
                    查找DeviceID
                </Button>
            </div>
            <div>{"用户UserID : " + @getAccountRes()}</div>
            <NamePassEditor {...@props} />
            <DeviceEditor   {...@props} />
        </div>

module.exports = App
