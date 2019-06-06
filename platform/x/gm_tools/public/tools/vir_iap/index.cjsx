antd = require 'antd'
React      = require 'react'
AccountKeyInput = require '../../common/account_input'
VirIAPTable     = require './vir_iap_table'
NewSendIAPModal = require './send_vir_iap'
VirTrueIAP      = require './vir_true_iap'

IAPApp = React.createClass
    getInitialState: () ->
        return {
            select_server : @props.curr_server
            player_to_send : ""

            loadstat : 0
        }

    enterLoading: () ->
        @setState {
            loadstat : 1
        }
        if @refs.iap_table?
            @refs.iap_table.Refersh(1)
        console.log @state

    leaveLoading: () ->
        @setState {
            loadstat : 2
        }

    isServerRight : () -> @state.select_server? and @state.select_server.serverName?

    isAccountRight : () -> @refs.accountin? and @refs.accountin.IsRight()

    getLoadingState : () ->
        if not @isServerRight() or not @isAccountRight()
            return "disabled"
        return if @state.loadstat is 1 then 'ant-btn-loading' else ''

    getIAPTable : () ->
        if @state.loadstat is 0
            return <div>点击按钮加载ID</div>

        if not @isServerRight()
            return <div> </div>

        if not @isAccountRight()
            return <div> </div>

        return <VirIAPTable
                    {...@props}
                    ref        = "iap_table"
                    server_id  = { @state.select_server.serverName  }
                    account_id = { @state.player_to_send }
                    on_loaded  = { @leaveLoading         } />

    getSendIAPModal : () ->
        if not @isServerRight()
            return <div>请选择服务器</div>

        if not @isAccountRight()
            return <div>请输入正确的用户ID</div>

        return <NewSendIAPModal
                    {...@props}
                    modal_name = "发送虚拟充值邮件"
                    server_id  = { @state.select_server.serverName  }
                    account_id = { @state.player_to_send }
                    onSend =     { @OnSend               } />

    getSendTrueIAPModal : () ->
        if not @isServerRight()
            return <div>请选择服务器</div>

        if not @isAccountRight()
            return <div>请输入正确的用户ID</div>

        return <VirTrueIAP
                    {...@props}
                    modal_name = "发送真实虚拟充值邮件"
                    server_id  = { @state.select_server.serverName  }
                    account_id = { @state.player_to_send }
                    onSend =     { @OnSend               } />



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

    render:() ->
        <div>
            <AccountKeyInput {...@props} ref = "accountin" can_cb = {@handleUserChange}/>
            {@getSendIAPModal()}
            {@getSendTrueIAPModal()}
            <antd.Button className={'ant-btn anticon-search ant-btn-primary ' + @getLoadingState()} onClick={@enterLoading}>
                加载用户虚拟充值记录
            </antd.Button>

            {@getIAPTable()}
        </div>

module.exports = IAPApp
