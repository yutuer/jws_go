antd        = require 'antd'
Api         = require '../api/api_ajax'
Modal       = antd.Modal
DatePicker  = antd.DatePicker
RangePicker = DatePicker.RangePicker
InputNumber = antd.InputNumber
RewardMaker = require '../../common/reward_maker'
SysRollNotice = require './notice'
TimeUtil    = require '../util/time'
React      = require 'react'

multiLang = require '../../common/multi_lang'
MultiLangModal = multiLang.MultiLangModal

NewModal = React.createClass
    getInitialState: () ->
        if @props.notice?
            return {
                loading: false,
                visible: false,
                notice : @props.notice
            }
        else
            n = (new Date()).getTime()
            return {
                loading: false,
                visible: false,
                notice : new SysRollNotice {
                    begintime    : n
                    endtime      : n
                    id       : 0
                    interval : 30
                    command : {
                        server:  "0:0"
                        params : ["", "", "", "0", "", "","", "",""]
                    }
                }
            }

    showModal : () ->
        @setState { visible: true }

    handleOk : () ->
        @setState { loading: true }
        @setState { loading: false, visible: false }
        @props.on_ok(@state.notice)

    handleCancel : () ->
        @setState { visible: false }

    handleStartStopChange : (value) ->
        console.log('From: ', value[0], ', to: ', value[1])
        notice = @state.notice
        notice.begin = parseInt(value[0].getTime(), 10)
        notice.end = parseInt(value[1].getTime(), 10)

        @setState { notice : notice }
        return

    handleLangChange : (value) ->
        notice = @state.notice
        notice.multi_lang = value.multi_lang
        notice.language = value.language
        notice.title = value.title
        notice.info = value.body
        @setState { notice : notice }

    handleIntervalChange  : (count) ->
        notice = @state.notice
        notice.interval = count
        @setState { notice : notice }
    getTimeToStr : (date, format) ->
        console.log date
        rs = TimeUtil.DateFormat((new Date(date * 1000)), format)
        console.log rs
        return rs

    render:() ->
            <Modal ref="modal"
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
                    <div>时间:</div>
                    <RangePicker showTime format="yyyy/MM/dd HH:mm:ss" onChange={@handleStartStopChange} />
                    <div>时间间隔(秒):</div>
                    <InputNumber
                        min={1}
                        max={1000000}
                        defaultValue={parseInt(@state.notice.interval, 10)}
                        onChange={@handleIntervalChange}
                        style={{width:85}} />
                </div>

                <div className="row-flex row-flex-middle row-flex-start">
                    <MultiLangModal
                        {...@props}
                            title = @state.notice.title
                            body =@state.notice.info
                            multi_lang = @state.notice.multi_lang
                            lang = @state.notice.language
                            handleChange={@handleLangChange}>
                    </MultiLangModal>
                </div>

            </Modal>

NewModalWithButton = React.createClass
    showModal : () ->
        @refs.sys_roll_notice_modal.showModal()

    handleOk : (notice) ->
        console.log notice
        loadFunc = () =>@props.on_loaded()
        notice.NewToServer id, @props.account_id, @props.curr_key, loadFunc  for id in @props.server_id


    render:() ->
        <div>
            <NewModal
                {...@props}
                ref = "sys_roll_notice_modal"
                on_ok = {@handleOk} />
            <button
                className="ant-btn ant-btn-primary"
                onClick={@showModal}>
                {@props.modal_name}
            </button>
        </div>

module.exports = {
    NewModal : NewModal
    NewModalWithButton : NewModalWithButton
}
