antd = require 'antd'
React      = require 'react'
ProfileQuery = require '../../scripts/gm_profile_query'

ProfileQuery = ProfileQuery.ProfileQuery

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


    getServerName : () ->
        if @state.select_server?
            return @state.select_server.name
        else
            return ""


    render:() ->
        <div>
            <ProfileQuery typ="Modify" />
        </div>

module.exports = App
