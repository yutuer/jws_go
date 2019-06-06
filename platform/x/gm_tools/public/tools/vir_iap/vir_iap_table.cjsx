antd    = require 'antd'
boot    = require 'react-bootstrap'
Api     = require '../api/api_ajax'
React   = require 'react'
Table = boot.Table

time_2_str = (t) ->
  d = new Date(t * 1000)
  return d.toString()

VirIAPTable = React.createClass
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

        api = new Api()
        api.Typ("getVirtualIAP")
           .ServerID(@props.server_id)
           .AccountID(@props.account_id)
           .Key(@props.curr_key)
           .Do (result) =>
                @setState {
                    mails : JSON.parse(result)
                    is_loading : false
                }
                console.log "on_loaded"
                @props.on_loaded()

    getRewardInf : (ids, cs) ->
        re = ""
        for idx, item_id of ids
            re = re + item_id + ":" + cs[idx] + ","
        return re

    isIAP : (v) ->
        if v.ItemId != null
            is_iap = v.ItemId[0] == "VI_HC_Buy"
            return [is_iap, v.Count[0], v.IsGetted]
        else
            return [0,0,0]

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
            api.Typ("delVirtualIAP")
               .ServerID(@props.server_id)
               .AccountID(@props.account_id)
               .Key(@props.curr_key)
               .Params(id)
               .Do (result) =>
                    console.log "on_deled"
                    @Refersh(1)

    getAllInfo : (data) ->
        if not data?
          return <div>UnKnown Info</div>

        re = []
        for k, v of data
          [is_iap, hc, is_get] = @isIAP(v)

          if is_iap
              re.push <tr key={k} >
                      <td>{v.Idx}</td>
                      <td>{time_2_str(v.TimeEnd)}</td>
                      <td>{hc}</td>
                      <td>{if is_get then "已领取" else "未领取" }</td>
                      <td>
                          <boot.Button bsStyle='danger'
                                      disabled = {is_get}
                                      onClick={@del(v.Idx)} >
                                      删除
                          </boot.Button>
                      </td>
                  </tr>
        return re

    getRewardList : () ->
        if @state.is_loading
          return <i className="anticon anticon-loading"></i>

        return <boot.Table striped bordered condensed hover>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>结束时间</th>
                    <th>奖励HC</th>
                    <th>是否已领取</th>
                    <th>删除</th>
                </tr>
            </thead>
            <tbody>
                { @getAllInfo(@state.mails) }
            </tbody>
        </boot.Table>

    render:() ->
        <div className="ant-form-inline">
            {@getRewardList()}
        </div>

module.exports = VirIAPTable
