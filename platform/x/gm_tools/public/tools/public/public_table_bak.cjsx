antd         = require 'antd'
boot         = require 'react-bootstrap'
Api          = require '../api/api_ajax'
NewSendModal = require './new_public'
TimeUtil     = require '../util/time'
Table        = boot.Table
React      = require 'react'
###
notices : [
    {
        id:23234234,
        game_id:0
        priority:3
        typ:1
        title:ddddd
        body:fdsfsdfdsfds
        begin:23432432
        end:32423423
        state:0,1
        language:zh-Hans
    }
]

###
PublicTable = React.createClass
    getInitialState: () ->
        return {
            notices    : {}
            is_loading : true
        }

    componentDidMount: () -> @Refersh(0)

    Refersh: (time_wait) ->
        @setState {
            is_loading : true
        }

        if time_wait? and time_wait > 0
            setTimeout @Refersh, time_wait*1000
            return

        api = new Api()
        api.Typ("getSysPublic")
           .ServerID(@props.server_id)
           .AccountID(@props.account_id)
           .Key(@props.curr_key)
           .Params(@props.version)
           .Do (result) =>
                @setState {
                    notices : JSON.parse(result)
                    is_loading : false
                }
                console.log "on_loaded"
                @props.on_loaded()

    getDelButton : (is_get, id) ->
            return <delButton id  = {id}
              key={id}
              on_deled   = { () => @Refersh(1) }
              disabled   = { is_get            }
              server_id  = { @props.server_id  }
              account_id = { @props.account_id } />

    del : (id) ->
        return () =>
            api = new Api()
            api.Typ("delSysPublic")
               .ServerID(@props.server_id)
               .AccountID(@props.account_id)
               .Key(@props.curr_key)
               .Params(id)
               .Do (result) =>
                    console.log "on_deled"
                    @Refersh(0.3)

    modSendModal : (id) ->
        return () =>
            ref_str = "SysPublicModal" + id
            @refs[ref_str].ShowModal()

    send : (i) ->
        console.log "state"
        console.log @state.notices[i]
        api = new Api()


        start = new Date(@state.notices[i].begin)
        year = start.getFullYear()
        month = (start.getMonth() + 1).toString()
        day = start.getDate().toString()
        hours = start.getHours().toString()
        minutes = start.getMinutes().toString()
        seconds = start.getSeconds().toString()
        if month.length == 1
          month = '0' + month
        if day.length == 1
          day = '0' + day
        if hours.length == 1
          hours = '0' + hours
        if minutes.length == 1
          minutes = '0' + minutes
        if seconds.length == 1
          seconds = '0' + seconds
        start_day_time = year + '/' + month + '/' + day + ' ' + hours + ':' + minutes + ':' + seconds

        end = new Date(@state.notices[i].end)
        year = end.getFullYear()
        month = (end.getMonth() + 1).toString()
        day = end.getDate().toString()
        hours = (end.getHours()).toString()
        minutes = end.getMinutes().toString()
        seconds = end.getSeconds().toString()
        if month.length == 1
            month = '0' + month;

        if day.length == 1
            day = '0' + day;

        if hours.length == 1
            hours = '0' + hours;

        if minutes.length == 1
            minutes = '0' + minutes;

        if seconds.length == 1
            seconds = '0' + seconds;

        end_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" + seconds;



        api.Typ("sendSysPublic")
           .ServerID(@props.server_id)
           .AccountID(i)
           .Key(@props.curr_key)
           .Params(
               @props.gid,
               @state.notices[i].priority,
               @state.notices[i].typ,
               start_day_time,
               end_day_time,
               @state.notices[i].state,
               @state.notices[i].title,
               @state.notices[i].body,
               @props.version,
               @state.notices[i].class,
               @state.notices[i].lang,
               @state.notices[i].multi_lang
           )
           .Do (result) =>
                @setState {
                    notices : JSON.parse(result)
                    is_loading : false
                }
                console.log "on_loaded"
                @Refersh(0.3)

    activityPublic : (id) ->
        return () =>
            states = @state.notices
            if states[id].state is "0" or states[id].state is 0
                states[id].state = "1"
            else
                states[id].state = "0"
            @setState {
                notices : states
            }
            @send id

    getPublicClass : (c) ->
        switch c
            when "Publics" then "通常"
            when "Maintaince" then "维护"
            when "Forceupdate" then "强制更新"

    getStateString : (v) ->
        if v.state is 0
            return "未发布"

        now_t = new Date()
        begin_t = new Date(v.begin)
        end_t = new Date(v.end)

        if begin_t <= now_t < end_t
            return "已发布"

        if now_t >= end_t
            return "已过期"

        if now_t < begin_t
            return "未到期"

    getAllInfo : (data) ->
        if not data?
          return <div>UnKnown Info</div>

        re = []
        for k, v of data
            console.log v
            console.log v.game_id
            console.log @props.gid
            re.push <tr key={v.id} >
                      <td>{v.id}</td>
                      <td>{@getPublicClass(v.class)}</td>
                      <td>{v.game_id}</td>
                      <td>{v.priority}</td>
                      <td>{v.typ}</td>
                      <td>{v.title}</td>
                      <td>{v.begin}</td>
                      <td>{v.end}</td>
                      <td>{@getStateString(v)}</td>
                      <td className='row-flex row-flex-middle row-flex-start'>
                          <NewSendModal.SysPublicModal
                                      {...@props}
                                      ref = {"SysPublicModal" + v.id}
                                      id  = {v.id}
                                      modal_name = "sds"
                                      key = {v.id}
                                      on_loaded  = { () => @Refersh(0.3) }
                                      priority   = { v.priority}
                                      typ        = { v.typ}
                                      start_time = { v.begin}
                                      end_time   = { v.end}
                                      state      = { v.state}
                                      title      = { v.title}
                                      body       = { v.body}
                                      class      = { v.class}
                                      lang       = { v.lang}
                                      multi_lang = { v.multi_lang}
                                      />
                          <boot.Button bsStyle='danger'
                                      disabled = {false}
                                      onClick={ @activityPublic(v.id) } >
                                      发布
                          </boot.Button>
                          <boot.Button bsStyle='danger'
                                      disabled = {false}
                                      onClick={ @modSendModal(v.id) } >
                                      修改
                          </boot.Button>
                          <boot.Button bsStyle='danger'
                                  disabled = {false}
                                  onClick={@del(v.id)} >
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
                    <th>种类</th>
                    <th>公告大区</th>
                    <th>优先级</th>
                    <th>类型</th>
                    <th>公告标题</th>
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

module.exports = PublicTable
