antd = require 'antd'
React      = require 'react'
Modal      = antd.Modal
DatePicker  = antd.DatePicker
RangePicker = DatePicker.RangePicker
Checkbox = antd.Checkbox
$          = require 'jquery'

RewardMaker = require '../../common/reward_maker'
AccountList = require '../../common/account_list'

multiLang = require '../../common/multi_lang'
MultiLangModal = multiLang.MultiLangModal

NewMailAllSendModal = React.createClass
    getInitialState: () ->
        return {
          loading: false,
          visible: false,

          title      : null,
          info       : null,
          start_time : null,
          end_time   : null,
          c_start_time : null,
          c_end_time   : null,
          ver        : null,
          rewards    : [],
          accounts    : [],
          send_all_ser: false,
          language   : "",
          multi_lang : "",
          profile    : @props.mail_name
        }

    showModal : () ->
        @setState { visible: true }

    handleOk : () ->
        @setState { loading: true }
        @setState { loading: false, visible: false }
        @send()

    handleCancel : () ->
        @setState { visible: false }

    send : () ->
        console.log @state

        item_id_ar = []
        count_ar = []

        for reward in @state.rewards
            item_id_ar.push reward.Id
            count_ar.push reward.Count

        h_time = @state.start_time
        year = h_time.getFullYear()
        month = (h_time.getMonth() + 1).toString()
        day = h_time.getDate().toString()
        hours = (h_time.getHours()).toString()
        minutes = h_time.getMinutes().toString()
        seconds = h_time.getSeconds().toString()
        if month.length == 1
         month = '0' + month
        if day.length == 1
         day = '0' + day
        if hours.length == 1
          hours = '0' + hours
        if minutes.length == 1
          minutes = '0' + minutes
        if seconds.length == 1
          seconds='0'+ seconds
        start_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" +seconds


        h_time = @state.end_time
        year = h_time.getFullYear()
        month = (h_time.getMonth() + 1).toString()
        day = h_time.getDate().toString()
        hours = (h_time.getHours()).toString()
        minutes = h_time.getMinutes().toString()
        seconds = h_time.getSeconds().toString()
        if month.length == 1
         month = '0' + month
        if day.length == 1
         day = '0' + day
        if hours.length == 1
          hours = '0' + hours
        if minutes.length == 1
          minutes = '0' + minutes
        if seconds.length == 1
          seconds='0'+ seconds
        end_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" +seconds
        ##将全部时间转换为字符串发送到后台




        ndata = {
            "Info"      : @state.info,
            "Title"     : @state.title,
            "ItemId"    : item_id_ar,
            "Count"     : count_ar,
            "TimeBegin" : parseInt(@state.start_time.getTime()   / 1000,10),
            "TimeBeginString" : start_day_time,
            "TimeEnd"   : parseInt(@state.end_time.getTime()   / 1000,10),
            "TimeEndString"   : end_day_time,
            "Idx"       : parseInt(new Date().getTime()        / 1000, 10),
            "CTimeBegin": 0,
            "CTimeEnd"  : 0,
            "Accounts"  : @state.accounts,
            "SendAllSer": @state.send_all_ser,
            "Lang" : JSON.stringify(@state.language),
            "MultiLang":@state.multi_lang
        }

        if @state.c_start_time?
            h_time = @state.c_start_time
            year = h_time.getFullYear()
            month = (h_time.getMonth() + 1).toString()
            day = h_time.getDate().toString()
            hours = (h_time.getHours()).toString()
            minutes = h_time.getMinutes().toString()
            seconds = h_time.getSeconds().toString()
            if month.length == 1
             month = '0' + month
            if day.length == 1
             day = '0' + day
            if hours.length == 1
              hours = '0' + hours
            if minutes.length == 1
              minutes = '0' + minutes
            if seconds.length == 1
              seconds='0'+ seconds
            c_start_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" +seconds
            ndata["CTimeBegin"] = parseInt(@state.c_start_time  / 1000,10)
            ndata["CTimeBeginString"] = c_start_day_time

        if @state.c_end_time?
            h_time = @state.c_end_time
            year = h_time.getFullYear()
            month = (h_time.getMonth() + 1).toString()
            day = h_time.getDate().toString()
            hours = (h_time.getHours()).toString()
            minutes = h_time.getMinutes().toString()
            seconds = h_time.getSeconds().toString()
            if month.length == 1
             month = '0' + month
            if day.length == 1
             day = '0' + day
            if hours.length == 1
              hours = '0' + hours
            if minutes.length == 1
              minutes = '0' + minutes
            if seconds.length == 1
              seconds='0'+ seconds
            c_end_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" +seconds
            ndata["CTimeEnd"] = parseInt(@state.c_end_time  / 1000,10)
            ndata["CTimeEndString"] = c_end_day_time


        if @state.ver? and @state.ver isnt ""
            ndata["Tag"] = JSON.stringify {"ver" : @state.ver}

        console.log(ndata)
        self = @
        $.ajax {
            url : "../api/v1/mail/" + @props.server_name + "/" +  @props.mail_name,
            dataType: "json",
            type: "POST",
            data: JSON.stringify(ndata),
            contentType: "application/json; charset=utf-8",
            complete: (xhr, status) ->
                rc = {
                    status: xhr.status
                    json: xhr.responseJSON
                }
                console.log rc
                self.props.onSend()
        }

    handleStartStopChange : (value) ->
        console.log('handleStartStopChange From: ', value[0], ', to: ', value[1])
        @setState {
            start_time : new Date(value[0].getTime())
            end_time   : new Date(value[1].getTime())
        }
        return
    handleCreatAccStartStopChange : (value) ->
        console.log('handleCreatAccStartStopChange From: ', value[0], ', to: ', value[1])
        if value[0]?
            @setState {
                c_start_time : new Date(value[0].getTime())
            }
        else
            @setState {
                c_start_time : null
            }

        if value[1]?
            @setState {
                c_end_time   : new Date(value[1].getTime())
            }
        else
            @setState {
                c_end_time   : null
            }

        return
    handleTitleChange     : (event) -> @setState { title      : event.target.value    }
    handleInfoChange      : (event) -> @setState { info       : event.target.value    }
    handleVerChange       : (event) -> @setState { ver        : event.target.value    }

    handleRewardChange : (crewards) ->
        @setState {
            rewards : crewards
        }
        console.log "handleRewardChange"
        console.log @state

    handleLangChange : (value) ->
        @setState {
          title      : value.title,
          info       : value.body,
          language   : value.language,
          multi_lang : value.multi_lang
        }

    onAccountListChange : (caccounts) ->
        @setState {
            accounts : caccounts
        }
        console.log "onAccountListChange"
        console.log @state

    onSendAllServer : (e) ->
        @setState {
            send_all_ser : e.target.checked
        }

    render:() ->
        <div>
            <button className="ant-btn ant-btn-primary" onClick={this.showModal}>
                {@props.modal_name}
            </button>
            <Modal ref="modal"
                visible={@state.visible}
                title={@props.modal_name} onOk={@handleOk} onCancel={@handleCancel}
                width=700
                footer={[
                    <button key="back" className="ant-btn" onClick={@handleCancel}>返 回</button>,
                    <button key="submit" className="ant-btn ant-btn-primary" onClick={@handleOk}>
                        提 交
                        <i className={'anticon anticon-loading'+(@state.loading?'':'hide')}></i>
                    </button>
                ]}>

                <div className="row-flex row-flex-middle row-flex-start">
                    <div>时间:</div>
                    <RangePicker showTime format="yyyy/MM/dd HH:mm" onChange={@handleStartStopChange} />
                </div>

                <div className="row-flex row-flex-middle row-flex-start">
                    <MultiLangModal handleChange={@handleLangChange}>
                    </MultiLangModal>
                </div>

                <p>奖励:（VI_HC类型的不能超过500）</p>
                <RewardMaker onChange={@handleRewardChange} />

                <div className="row-flex row-flex-middle row-flex-start">
                    <div>邮件限定版本(不填则省略):</div>
                    <input className="ant-input " onChange={@handleVerChange}></input>
                </div>

                <div>
                    <div>玩家注册时间限制(只有在此时间内注册的玩家能见到邮件):</div>
                    <RangePicker showTime format="yyyy/MM/dd HH:mm" onChange={@handleCreatAccStartStopChange} />
                    <div>玩家ID限制(只有在列表中的玩家能见到邮件):</div>
                    <AccountList onChange={@onAccountListChange}/>
                </div>

                <div>
                    <Checkbox onChange={@onSendAllServer}> 
                        发给全区所有服务器 
                    </Checkbox>
                     发给全区所有服务器 （需要勾选一个此大区的服务器）
                </div>
            </Modal>
        </div>

module.exports = NewMailAllSendModal
