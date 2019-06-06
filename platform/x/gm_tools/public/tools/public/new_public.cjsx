antd        = require 'antd'
Api         = require '../api/api_ajax'
Modal       = antd.Modal
DatePicker  = antd.DatePicker
RangePicker = DatePicker.RangePicker
InputNumber = antd.InputNumber
RewardMaker = require '../../common/reward_maker'
TimeUtil    = require '../util/time'
React      = require 'react'
Select = antd.Select

multiLang = require '../../common/multi_lang'
LanguagePicker = multiLang.LanguagePicker
MultiLangPicker = multiLang.MultiLangPicker
langMap = multiLang.langMap
MultiLangModal = multiLang.MultiLangModal

Picker = React.createClass
  handleChange : (data) ->
    @props.handleChange(data)

  render:() ->
      <Select defaultValue = {@props.defaultValue} style={{width:85}} onChange={@handleChange}>
          <Option value={"Publics"}>通常</Option>
          <Option value={"Maintaince"}>维护</Option>
          <Option value={"Forceupdate"}>强制更新</Option>
      </Select>

ClassPicker = React.createClass
  handleChange : (data) ->
    @props.handleChange(data)

  render:() ->
      <Select defaultValue = {@props.defaultValue} style={{width:85}} onChange={@handleChange}>
          <Option value={0 }>活动</Option>
          <Option value={1 }>促销</Option>
          <Option value={2 }>最新</Option>
          <Option value={3 }>福利</Option>
          <Option value={4 }>节日</Option>
          <Option value={5 }>充值</Option>
          <Option value={6 }>消耗</Option>
          <Option value={7 }>热卖</Option>
          <Option value={8 }>神将</Option>
          <Option value={9 }>神装</Option>
          <Option value={10}>New</Option>
          <Option value={11}>Hot</Option>
          <Option value={12}>公告</Option>
          <Option value={13}>停机</Option>
          <Option value={14}>更新</Option>
          <Option value={15}>版本</Option>
      </Select>

SysPublicModal = React.createClass
    getInitialState: () ->
        start =  if @props.start_time
                     new Date(@props.start_time * 1000)
                 else
                     new Date()
        end   =  if @props.end_time
                     new Date(@props.end_time * 1000)
                 else
                     new Date()

        propsLang = @props.lang
        tempLang = {}
        if !@IsJsonString(propsLang)
            for key of langMap
                tempLang[key] = {title:"", body:""}
        else
            tempLang = JSON.parse(propsLang)

        s = {
          loading    : false,
          visible    : false,
          title      : @props.title,
          body       : @props.body,
          start_time : start,
          end_time   : end,
          gid        : @props.gid,
          priority   : @props.priority,
          typ        : @props.typ,
          class      : @props.class,
          language   : tempLang,
          multi_lang : @props.multi_lang
        }
        return s

    IsJsonString : (jsonStr) ->
        if jsonStr == "" || !jsonStr
            return false
        try
            JSON.parse jsonStr
            return true
        catch err
            return false

    ShowModal : () ->
        @setState { visible: true }

    handleOk : () ->
        @setState { loading: true }
        @setState { loading: false, visible: false }
        @send()

    handleCancel : () ->
        @setState { visible: false }

    send : () ->
        console.log @state
        console.log @props
        api = new Api()

        start = new Date(@state.start_time)
        year = start.getFullYear()
        month = (start.getMonth() + 1).toString()
        day = start.getDate().toString()
        hours = (start.getHours()).toString()
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
          seconds='0'+ seconds
        start_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" +seconds

        end = new Date(@state.end_time)
        year = end.getFullYear()
        month = (end.getMonth() + 1).toString()
        day = end.getDate().toString()
        hours = (end.getHours()).toString()
        minutes = end.getMinutes().toString()
        seconds = end.getSeconds().toString()
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
        end_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" + seconds

        api.Typ("sendSysPublic")
           .ServerID(@props.server_id)
           .AccountID(@props.id)
           .Key(@props.curr_key)
           .Params(
               @props.gid,
               @state.priority,
               @state.typ,
               start_day_time,
               end_day_time,
               "0",
               @state.title,
               @state.body,
               @props.version,
               @state.class,
               JSON.stringify(@state.language),
               @state.multi_lang
           )
           .Do (result) =>
                @setState {
                    notices : JSON.parse(result)
                    is_loading : false
                }
                console.log "on_loaded"
                @props.on_loaded()

    getTimeToStr : (date, format) ->
        console.log date
        rs = TimeUtil.DateFormat(date, format)
        console.log rs
        return rs

    handleStartStopChange : (value) ->
        console.log('From: ', value[0], ', to: ', value[1])
        @setState {
            start_time : value[0]
            end_time   : value[1]
        }
        return

    handlePriorityChange  : (event) -> @setState { priority   : parseInt(event.target.value, 10)    }
    handleTypChange       : (value) -> @setState { typ        : parseInt(value, 10)    }
    handleClassChange     : (value) -> @setState { class      : value    }
    handleLangChange : (value) ->
        @setState {
          title      : value.title,
          body       : value.body,
          language   : value.language,
          multi_lang : value.multi_lang
        }

    render:() ->
            <Modal
                {...@props}
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
                    <div>种类:</div>
                    <Picker  defaultValue = {@state.class} handleChange={@handleClassChange}></Picker>
                    <div>优先级:</div>
                    <input className="ant-input" value = {@state.priority} style={{width:70}} onChange={@handlePriorityChange}></input>
                    <div>类型:</div>
                    <ClassPicker  defaultValue = {@state.typ} handleChange={@handleTypChange}></ClassPicker>
                </div>

                <div className="row-flex row-flex-middle row-flex-start">
                    <div>时间:</div>
                    <RangePicker showTime format="yyyy/MM/dd HH:mm" onChange={@handleStartStopChange} />
                </div>

                <div className="row-flex row-flex-middle row-flex-start">
                    <MultiLangModal
                        {...@props}
                            title = @state.title
                            body = @state.body
                            multi_lang = @state.multi_lang
                            lang = @state.language
                            handleChange={@handleLangChange}>
                    </MultiLangModal>
                </div>

            </Modal>


NewModal = React.createClass
    getInitialState: () -> {}
    render:() ->
        <div>
            <SysPublicModal
                {...@props}
                id="null"
                ref="modal"
            />
            <button className="ant-btn ant-btn-primary" onClick={() => @refs.modal.ShowModal()}>
                {@props.modal_name}
            </button>
        </div>

module.exports = {
    NewModal : NewModal
    SysPublicModal : SysPublicModal
}
