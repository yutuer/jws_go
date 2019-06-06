antd     = require 'antd'
boot     = require 'react-bootstrap'
CSVInput = require '../../common/csv_input'
Api      = require '../api/api_ajax'
React      = require 'react'
Input    = boot.Input
Table    = boot.Table
Button   = antd.Button

App = React.createClass
    getInitialState: () ->
        return {
            select_server : @props.curr_server
            send : []
            res  : []
            content : ""
        }

    isServerRight  : () -> @state.select_server? and @state.select_server.serverName?

    getLoadingState : () ->
        if not @isServerRight()
            return "disabled"
        return ''

    handleServerChange: (data) ->
        console.log data
        @setState {
            select_server : data
        }

    handleChange : (v, str, is_right) ->
        console.log v

        nick_names = v.map (a) -> a[0]
        console.log nick_names
        @setState {
            send: nick_names
            content : str
        }

    getServerName : () ->
        if @state.select_server?
            return @state.select_server.name
        else
            return ""

    getNick : () ->
        api = new Api()
        console.log @state.send
        api.Typ("getInfoByNickName")
           .ServerID(@state.select_server.serverName)
           .AccountID("AID")
           .Key(@props.curr_key)
           .ParamArray(@state.send)
           .Do (result) =>
                console.log "onSend"
                console.log result
                @setState {
                    res : JSON.parse(result)
                }

    getRes : () ->
        <Table  {...@props} striped bordered condensed hover >
            <thead>
                <tr>
                    <th>昵称</th>
                    <th>ID</th>
                </tr>
            </thead>
            <tbody>
                {@state.res.map (v) ->
                    <tr key={v.name} >
                        <td> {v.name} </td>
                        <td> {v.pid}  </td>
                    </tr>}
            </tbody>
        </Table>

    render:() ->
        <div>
            <CSVInput {...@props} title="请输入昵称" value={@state.content} ref = "accountin" can_cb = {@handleChange}/>
            <Button className={'ant-btn anticon-search ant-btn-primary ' + @getLoadingState()} onClick={@getNick}>
                查询
            </Button>
            { @getRes() }
        </div>

module.exports = App
