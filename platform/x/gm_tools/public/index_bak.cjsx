antd       = require 'antd'
ReactDOM   = require 'react-dom'
React      = require 'react'

require 'antd/lib/index.css'
require 'bootstrap/dist/css/bootstrap.min.css'
require 'jsoneditor/dist/jsoneditor.min.css'

Mail           = require './tools/mail_all/index'
VirIAP         = require './tools/vir_iap/index'
SysRollNotice  = require './tools/sys_roll_notice/index'
SysPublic      = require './tools/public/index'
BanAccount     = require './tools/ban_account/index'
Login          = require './login/login'
ProfileQuery   = require './tools/profile_query/index'
GagAccount     = require './tools/gag_account/index'
NickName       = require './tools/nickname/index'
ChannelUrl     = require './tools/channelurl/index'
IAPQuery       = require './tools/iap_query/index'
ShardMng       = require './tools/shard_mng/index'
ShardShowState = require './tools/shard_show_state/index'
ActValid       = require './tools/act_valid/index'
DataVer        = require './tools/data_ver/index'
RankDel        = require './tools/rank_del/index'
ProfileTrans   = require './tools/trans_account/index'
DeviceTrans    = require './tools/trans_device/index'

ReactCookie = require 'react-cookie'

func_commands = {
    "昵称查询"   : NickName
    "单人邮件"   : Mail.MailOne
    "全服邮件"   : Mail.MailAll
    "批量邮件"   : Mail.MailBatch
    "虚拟充值"   : VirIAP
    "跑马灯公告" : SysRollNotice
    "公告"      : SysPublic
    "封禁账号"   : BanAccount
    "禁言账号"   : GagAccount
    "账号查询"   : ProfileQuery
    "账号转移"   : ProfileTrans
    "渠道链接"   : ChannelUrl
    "支付查询"   : IAPQuery
    "Shard管理"    : ShardMng
    "Shard状态标签" : ShardShowState
    "活动开关"   : ActValid
    "数据版本"   : DataVer
    "删除排行榜" : RankDel
    "设备转移": DeviceTrans
}

activity_commands = {

}

server_commands = {
}

auth_commands = {

}

Siders  = require './common/siders'

Apps = React.createClass {
    getInitialState : () ->
        {
            curr : null
            curr_title : ""
            curr_key : ReactCookie.load('taihe-gm-tool-key')
        }

    handleServerChange : (e) ->
        @setState {
            curr_server : e
        }
        curr = @getCurrApp()
        console.log "handleServerChange ", @refs.curr_app, @refs.curr_app.handleServerChange
        @refs.curr_app.handleServerChange(e)

    handleKeyChange : (e) ->
        return if e is ""

        @setState {
            curr_key : e
        }
        ReactCookie.save 'taihe-gm-tool-key', e, {
            path    : '/'
            maxAge  :  3600
        }

        curr = @getCurrApp()
        console.log "handleKeyChange ", @state.curr_key

    handleSiderClick : (e) ->
        console.log("handleServerChange", func_commands[e.key])
        @setState {
            curr : func_commands[e.key]
            curr_title : e.key
        }

    handleLogout : () ->
        @setState {
            curr_key : ""
        }
        ReactCookie.save 'taihe-gm-tool-key', "", {
            path    : '/'
        }


    getCurrApp : () ->
        Curr = @state.curr
        console.log @state.curr_key
        if not @state.curr_key? or @state.curr_key is ""
            return <Login onChange = {@handleKeyChange}/>

        if Curr?
            return <Curr
                        ref         = "curr_app"
                        curr_server = {@state.curr_server}
                        curr_key    = {@state.curr_key} />
        else
            return <div>请选择功能</div>



    render : () ->
          <div className="row-flex row-flex-start">
              <div>
                  <Siders.Sider
                    handleClick={@handleSiderClick}
                    func_commands={func_commands}
                    activity_commands={activity_commands}
                    server_commands={server_commands}
                    auth_commands={auth_commands} />
              </div>
              <div className = "col-offset-1">
                  <Siders.StateMenu
                    handleServerChange={@handleServerChange}
                    handleLogout={@handleLogout}
                    className="row" />
                  <h3 type="info">{@state.curr_title}</h3>
                  <hr />
                  {@getCurrApp()}
                  <hr />
              </div>
          </div>
}
ReactDOM.render <Apps />, document.getElementById('react-content')
