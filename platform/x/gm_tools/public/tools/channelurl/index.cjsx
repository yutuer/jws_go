CSVInput = require '../../common/csv_input'
antd = require 'antd'
notification = antd.Notification
Api      = require '../api/api_ajax'
React      = require 'react'

openNotificationWithIcon = (type, title, msg) ->
    notification[type] {
        message: title,
        description: msg
        duration: 5
    }

App = React.createClass
    getInitialState: () ->
            return {
                res  : []
                csv  : ""
            }

    componentDidMount: () ->
        api = new Api()
        res = []
        api.Typ("getChannelUrl")
           .ServerID("nil")
           .AccountID("nil")
           .Key(@props.curr_key)
           .ParamArray([])
           .Do (result) =>
               console.log "onSend"
               console.log result
               res = JSON.parse(result)

               if res["ret"] != "ok"
                  openNotificationWithIcon("error", "Error", res["ret"])
               else
                  @setState {
                       res : res
                       csv : res["csv"]
                  }

    handleChange : (v, str, is_right) ->
        @setState {
             csv : str
        }
    
    onUpload : () ->
        api = new Api()
        res = []
        api.Typ("updateChannelUrl")
           .ServerID("nil")
           .AccountID("nil")
           .Key(@props.curr_key)
           .ParamArray([@state.csv])
           .Do (result) =>
               console.log "onSend"
               console.log result

    render:() ->
        <div>
            <CSVInput {...@props} title="客户端更新URL" value={@state.csv} ref = "accountin" can_cb = {@handleChange}/>
            <button className="ant-btn ant-btn-primary" onClick={@onUpload}>
                确定并上传
            </button>
        </div>


module.exports = App