antd          = require 'antd'
boot          = require 'react-bootstrap'
Api           = require '../api/api_ajax'
TimeUtil      = require '../util/time'
SysRollNotice = require './notice'
Table         = boot.Table
NewSendModal    = require './new_sys_roll_notice'
React      = require 'react'
NoticeTable = React.createClass
    getInitialState: () ->
        return {
            notices    : []
            is_loading : true
        }

    Refersh: (time_wait, callback) ->
        @setState {
            is_loading : true
        }

        if time_wait? and time_wait > 0
            setTimeout @Refersh.bind(@, null, callback), 1
            return

        api = new Api()
        all_notice = []
        count = 0

        (api.Typ("getSysRollNotice")
           .ServerID(server_id)
           .AccountID(@props.account_id)
           .Key(@props.curr_key)
           .Do (result) =>
                count++
                console.log result
                for k,v of JSON.parse(result)
                    all_notice[v.id] = new SysRollNotice(v)
                if count == @props.server_id.length
                    @setState {
                        notices    : all_notice
                        is_loading : false
                    }
                    @props.on_loaded() if @props.on_loaded?
                    callback()         if callback?) for server_id in @props.server_id




    getDelButton : (is_get, id, server_id) ->
            return <delButton id  = {id}
              key={id}
              on_deled   = { () => @Refersh(1) }
              disabled   = { is_get            }
              server_id  = { server_id  }
              account_id = { @props.account_id } />

    del : (id, server_id) ->
        return () =>
            api = new Api()
            api.Typ("delSysRollNotice")
               .ServerID(server_id)
               .AccountID(@props.account_id)
               .Key(@props.curr_key)
               .Params(id)
               .Do (result) =>
                    console.log "on_deled"
                    @Refersh(0.01)

    activityPublic : (id, server_id) ->
        return () =>
            states = @state.notices
            s = states[id]
            s.ChangeState()
            states[id] = s
            @setState {
                notices : states
            }
            s.UpdateToServer server_id, @props.account_id, @props.curr_key,
                () => @Refersh 0.01

    mod : (id) ->
        return () =>
            @refs["new_mod_" + id].showModal()

    handleModOk : (notice) ->
        console.log "handleModOk"
        console.log notice
        refresh =  () => @Refersh 0.01
        notice.UpdateToServer notice.server_id, @props.account_id, @props.curr_key, refresh


    getStateString : (v) ->
        if v.state is 0
            return "未发布"

        now_t = new Date()
        v_begin_t = new Date(v.json.command.params[0])
        v_end_t = new Date(v.json.command.params[1])

        if v_begin_t <= now_t < v_end_t
            return "已发布"
        if now_t >= v_end_t
            return "已过期"
        if now_t < v_begin_t
            return "未到期"


    getAllInfo : (data) ->
        if not data?
          return <div>UnKnown Info</div>

        re = []
        for k, v of data
            re.push <tr key={v.id} >
                      <td>{v.id}</td>
                      <td>{v.server_id}</td>
                      <td>{v.interval}</td>
                      <td>{v.title}</td>
                      <td>{v.json.command.params[0]}</td>
                      <td>{v.json.command.params[1]}</td>
                      <td>{v.State()}</td>
                      <td className='row-flex row-flex-middle row-flex-start'>
                          <NewSendModal.NewModal
                                {...@props}
                                ref = {"new_mod_" + v.id}
                                notice = {v}
                                server_id = [v.server_id]
                                on_ok = {@handleModOk} />
                          <boot.Button bsStyle='success'
                                      disabled = {false}
                                      onClick={@activityPublic(v.id, v.server_id)} >
                                      发送
                          </boot.Button>
                          <boot.Button bsStyle='info'
                                      disabled = {false}
                                      onClick={@mod(v.id)} >
                                      修改
                          </boot.Button>
                          <boot.Button bsStyle='danger'
                                  disabled = {false}
                                  onClick={@del(v.id, v.server_id)} >
                                  删除
                          </boot.Button>
                      </td>
                  </tr>
        return re

    getRewardList : () ->
        if @state.is_loading
          return <i className="anticon anticon-loading"></i>

        return <boot.Table  {...@props} striped bordered condensed hover>
            <thead>
                <tr>
                    <th>公告ID</th>
                    <th>公告大区</th>
                    <th>循环间隔(秒)</th>
                    <th>公告注释</th>
                    <th>开始时间</th>
                    <th>结束时间</th>
                    <th>状态</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody>
                { @getAllInfo(@state.notices) }
            </tbody>
        </boot.Table>

    render:() ->
        <div {...@props} className="ant-form-inline">
            {@getRewardList()}
        </div>

module.exports = NoticeTable
