antd = require 'antd'
React      = require 'react'
AccountKeyInput = require '../../common/account_input'
NoticeTable     = require './public_table'
NewSendModal    = require './new_public'
EndpointModal   = require './endpoint'
Api             = require '../api/api_ajax'

SysPublic = React.createClass
    getInitialState: () ->
        return {
            select_server : @props.curr_server
            player_to_send : ""
            version : ""
            gid : ""

            loadstat : {}
        }

    enterLoading: (t, cb) ->
        loadstat = @state.loadstat
        loadstat[t] = 1
        @setState {
            loadstat : loadstat
        }
        cb()
        return

    leaveLoading: (t) ->
        loadstat = @state.loadstat
        loadstat[t] = 2
        @setState {
            loadstat : loadstat
        }

    getLoadingState : (t) -> if @state.loadstat[t] is 1 then 'ant-btn-loading' else ''

    getTable : () ->
        if @state.loadstat is 0
            return <div>点击按钮加载ID</div>

        return <NoticeTable
                    {...@props}
                    ref        = "notice_table"
                    server_id  = "sid"
                    account_id = "ac"
                    version    = { @state.version                 }
                    gid        = { @state.gid                     }
                    on_loaded  = { () => @leaveLoading("table")   } />

    getSendModal : () ->
        return <NewSendModal.NewModal
                    {...@props}
                    ref        = "notice_send"
                    modal_name = "添加公告"
                    server_id  = "sid"
                    account_id = "ac"
                    version    = { @state.version        }
                    gid        = { @state.gid            }
                    on_loaded  = { @OnSend               } />

    getMkEndpointModal : () ->
        return <EndpointModal
                    {...@props}
                    ref        = "endpoint"
                    modal_name = "修改Endpoint"
                    server_id  = "sid"
                    account_id = "ac"
                    version    = { @state.version        }
                    gid        = { @state.gid            }
                    on_loaded  = { ()->null              } />

    handleVersionChange: (event) ->
        @setState {
            version : event.target.value
        }
        console.log(@state)
        @OnSend()

    handleGidChange: (event) ->
        @setState {
            gid : event.target.value
        }
        console.log(@state)
        @OnSend()

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
        if @refs.notice_table?
            @refs.notice_table.Refersh(0)
        if @refs.endpoint?
            @refs.endpoint.Refersh(0)

    getServerName : () ->
        if @state.select_server?
            return @state.select_server.name
        else
            return ""

    updateTable : () ->
        if @refs.notice_table?
            @refs.notice_table.Refersh(1)
        else
            @leaveLoading("table")
        console.log @state

    releasePublic : () ->
        api = new Api()
        src = "releaseSysPublic"
        src = "releaseSysPublicToRelease" if @state.gid >= 200
        api.Typ(src)
           .ServerID("s")
           .AccountID("a")
           .Key(@props.curr_key)
           .Params(
               @state.gid,
               @state.version,
           )
           .Do (result) =>
                @leaveLoading("release")

    uploadKS3 : () ->
        api = new Api()
        src = "uploadKS3"
        api.Typ(src)
           .ServerID("s")
           .AccountID("a")
           .Key(@props.curr_key)
           .Params(
               @state.gid,
               @state.version,
           )
           .Do (result) =>
                @leaveLoading("uploadKS3")

    render:() ->
        <div>
            <div className="row-flex row-flex-middle row-flex-start">
                <div>大区:</div>
                <input className="ant-input" style={{width:100}} onChange={@handleGidChange}></input>
                <div>版本:</div>
                <input className="ant-input" style={{width:100}} onChange={@handleVersionChange}></input>
                <antd.Button
                    className = {'ant-btn ant-btn-primary ' + @getLoadingState("table")}
                    onClick   = { () => @enterLoading("table", @updateTable.bind(@)) } >
                    加载公告
                </antd.Button>
                <antd.Button
                    className = {'ant-btn ant-btn-primary ' + @getLoadingState("release")}
                    onClick   = { () => @enterLoading("release", @releasePublic.bind(@)) } >
                    发布公告
                </antd.Button>
                {@getSendModal()}
                {@getMkEndpointModal()}
                <antd.Button
                    className = {'ant-btn ant-btn-primary ' + @getLoadingState("release")}}
                    onClick   = { () => @enterLoading("uploadKS3", @uploadKS3.bind(@)) }>
                    上传金山云
                </antd.Button>
            </div>
            <hr />
            {@getTable()}
        </div>

module.exports = SysPublic
