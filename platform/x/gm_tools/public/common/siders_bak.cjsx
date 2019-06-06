antd       = require 'antd'
Menu       = antd.Menu
SubMenu    = Menu.SubMenu
React      = require 'react'

Sider = React.createClass {
    getInitialState : () ->
        return {
          current: '1'
        }

    handleClick : (e) ->
        console.log('click ', e)
        @setState {
          current: e.key
        }
        @props.handleClick e

    mkMenu : (k, v) ->
        <Menu.Item key={k}>{k}</Menu.Item>

    mkFuncCommands: () ->
        res = []
        for k, v of @props.func_commands
            res.push @mkMenu k, v
        console.log(res)
        console.log(@props.func_commands)
        return res
    mkActivityCommands: () ->
        res = []
        for k, v of @props.activity_commands
            res.push @mkMenu k, v
        return res
    mkServerCommands: () ->
        res = []
        for k, v of @props.server_commands
            res.push @mkMenu k, v
        return res
    mkAuthCommands: () ->
        res = []
        for k, v of @props.auth_commands
            @mkMenu k, v
        return res

    render : () ->
        <Menu onClick={@handleClick}
              defaultOpenKeys={['sub1', 'sub2', 'sub4', 'sub5']}
              mode="inline">
              <SubMenu key="sub1" title={<span><i className="anticon anticon-mail"></i><span> 功能管理 </span></span>}>
                    {@mkFuncCommands()}
              </SubMenu>
              <SubMenu key="sub2" title={<span><i className="anticon anticon-appstore"></i><span> 活动管理 </span></span>}>
                    {@mkActivityCommands()}
              </SubMenu>
              <SubMenu key="sub4" title={<span><i className="anticon anticon-laptop"></i><span> 服务器同步 </span></span>}>
                    {@mkServerCommands()}
              </SubMenu>
              <SubMenu key="sub5" title={<span><i className="anticon anticon-setting"></i><span> 账号权限管理 </span></span>}>
                    {@mkAuthCommands()}
              </SubMenu>
        </Menu>
}

ServerSelectors      = require '../common/server_selector'

ServerSelector = ServerSelectors.ServerSelector
ServerPicker   = ServerSelectors.ServerPicker

StateMenu = React.createClass {
    getInitialState : () ->
        {
            current: 'mail'
        }
    handleClick : (e) ->
        console.log('click ', e);
        @setState {
            current: e.key
        }
        @props.handleLogout() if e.key is "logout"

    handleServerChange : (e) ->
        @props.handleServerChange(e)

    render : () ->
        <Menu
            onClick={@handleClick}
            selectedKeys={[]}
            mode="horizontal">
            <Menu.Item key="server">
                <i className="anticon anticon-mail"></i>选择服务器
                <ServerSelector handleChange={@handleServerChange} />
            </Menu.Item>
            <Menu.Item key="logout">
                <i className="anticon anticon-mail"></i>登出
            </Menu.Item>
        </Menu>
}

module.exports = {
    Sider       : Sider
    StateMenu   : StateMenu
}
