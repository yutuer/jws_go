antd    = require 'antd'
boot    = require 'react-bootstrap'
Loading = require 'react-loading'
React   = require 'react'
Table   = boot.Table
Api     = require '../api/api_ajax'
Label   = boot.Label
$       = require 'jquery'

time_2_str = (t) ->
  d = new Date(t * 1000)
  res = d.getFullYear() + '-' +(d.getMonth() + 1)+'-'+d.getDate()+
        ' '+d.getHours() + ':'+d.getMinutes() + ':'+ d.getSeconds()
  return res

tag_2_str = (t) ->
    if !t 
       return ''
    temp = JSON.parse(t)
    return ["IAPID: ",temp.go,"count: ",temp.a]

MailList = React.createClass
    getInitialState: () ->
        return {
            mails      : []
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

        $.get "../api/v1/mail/"+ @props.server_name + "/" +  @props.mail_name, (result) =>
            console.log result
            @setState {
                mails : JSON.parse(result)
                is_loading : false
            }

    getRewardInf : (ids, cs) ->
        re = []
        for idx, item_id of ids
            re.push <tr key={idx} >
                    <td>{item_id}</td>
                    <td>{cs[idx]}</td>
                </tr>

        return <Table striped bordered condensed hover>
            <thead>
                <tr>
                    <th>ItemID</th>
                    <th>数量</th>
                </tr>
            </thead>
            <tbody>
                { re }
            </tbody>
        </Table>

    del : (id) ->
        return () =>
            api = new Api()
            api.Typ("delMail")
               .ServerID(@props.server_name)
               .AccountID(@props.mail_name)
               .Key(@props.curr_key)
               .Params(id)
               .Do (result) =>
                    console.log "on_deled " + id
                    @Refersh(0.3)

    getStateLabel : (v) ->
        if @props.is_personal
            if v.IsRead
                return <Label bsStyle="warning">已阅读</Label>
            if v.IsGetted
                return <Label bsStyle="default">已获取</Label>
            return <Label bsStyle="success">未获取</Label>
        else
            begin_t = new Date(v.TimeBeginString)
            end_t = new Date(v.TimeEndString)
            if begin_t <= now_t < end_t

            if begin_t <= now_t < end_t
                return <Label bsStyle="success">已发布</Label>

            if now_t >= end_t
                return <Label bsStyle="default">已过期</Label>

            if now_t < begin_t
                return <Label bsStyle="warning">未到期</Label>

    getAllInfo : (data) ->
        if not data?
          return <div>UnKnown Info</div>

        re = []
        for k, v of data
            if v.Param == null
              param0 = "gm"
              param1 = "gm"
            else
              param0 = v.Param[0]
              param1 = v.Param[1]
            re.push <tr key={k} >
                    <td>{v.Idx}</td>
                    <td>{param0}</td>
                    <td>{param1}</td>
                    <td>{time_2_str(v.TimeBegin)}</td>
                    <td>{time_2_str(v.TimeEnd)}</td>
                    <td>{@getStateLabel(v)}</td>
                    <td>{@getRewardInf(v.ItemId, v.Count)}</td>
                    <td>{tag_2_str(v.Tag)}</td>
                    <td className='row-flex row-flex-middle row-flex-start'>
                        <boot.Button bsStyle='danger'
                                      disabled = {false}
                                      onClick={ @del(v.Idx)} >
                                      删除
                        </boot.Button>
                    </td>
              </tr>
        return re

    getRewardList : () ->
        if @state.is_loading
          return <Loading type='spin' height="200" width="200" color='#61dafb' />

        return <Table striped bordered condensed hover>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>标题</th>
                    <th>信息</th>
                    <th>开始时间</th>
                    <th>结束时间</th>
                    <th>状态</th>
                    <th>奖励</th>
                    <th>奖励信息</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody>
                { @getAllInfo(@state.mails) }
            </tbody>
        </Table>

    render:() ->
        <div className="ant-form-inline">
            {@getRewardList()}
        </div>

BatchMailResult = React.createClass
    ResultMap: {"1": "时间格式错误", "2":"奖励格式错误", "3":"ID必须是整数", "4":"不存在的服务器", "5":"服务器发送失败", "6":"uid错误"}

    getInitialState: () ->
        return {
            count : 0
            mails : []
        }

    Refresh: (data) ->
        @setState {
            mails : data.FailArr
            count : data.SuccessCount
        }

    getAllInfo : (data) ->
        if not data?
          return <div>UnKnown Info</div>

        re = []
        for k, v of data
          re.push <tr key={k} >
                    <td>{v.Id}</td>
                    <td><Label bsStyle="danger">失败</Label></td>
                    <td>{@ResultMap[v.FailCode]}</td>
                </tr>
        return re

    getResultList : () ->
        if @state.is_loading
          return <Loading type='spin' height="200" width="200" color='#61dafb' />

        return <Table striped bordered condensed hover>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>结果</th>
                    <th>信息</th>
                </tr>
            </thead>
            <tbody>
                { @getAllInfo(@state.mails) }
            </tbody>
        </Table>

    render:() ->
        <div className="ant-form-inline">
            <div>
                <Label bsStyle="success">成功：{@state.count}</Label>
                <Label bsStyle="danger">失败：{Object.keys(@state.mails).length}</Label>
            </div>
            <hr/>
            {@getResultList()}
        </div>

module.exports = {
    "MailList": MailList
    "BatchMailResult": BatchMailResult
}
