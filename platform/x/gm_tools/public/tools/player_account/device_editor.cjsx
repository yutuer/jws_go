antd   = require 'antd'
React  = require 'react'
Modal  = antd.Modal
Api    = require '../api/api_ajax'

DeviceCreater = React.createClass
    getInitialState: () ->
        return {
          loading: false,
          visible: false,

          user_id    : "",
          name       : "",
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
        api = new Api()
        console.log @state.send
        api.Typ("createDeviceForUID")
           .ServerID("SID")
           .AccountID(@state.user_id)
           .Key(@props.curr_key)
           .Params(@state.name)
           .Do (result) =>
                console.log "on createDeviceForUID"
                console.log result

    handleUserIDChange   : (event) -> @setState { user_id : event.target.value }
    handleDeviceIDChange : (event) -> @setState { name    : event.target.value }

    render:() ->
        <div>
            <button className="ant-btn ant-btn-primary" onClick={@showModal}>
                绑定UID到DeviceID
            </button>
            <Modal ref="modal"
                visible={@state.visible}
                title="绑定DeviceID到用户名" onOk={@handleOk} onCancel={@handleCancel}
                width=700
                footer={[
                    <button key="back" className="ant-btn" onClick={@handleCancel}>返 回</button>,
                    <button key="submit" className="ant-btn ant-btn-primary" onClick={@handleOk}>
                        提 交
                        <i className={'anticon anticon-loading'+(@state.loading?'':'hide')}></i>
                    </button>
                ]}>
                <div>注意:绑定UID到DeviceID之前要查询DeviceID是否存在, 如果存在请备份UID记录, 当DeviceID已存在时会覆盖原有的</div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>UserID(xx-xx-xx):</div>
                    <input className="ant-input " onChange={@handleUserIDChange}></input>
                </div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>DeviceID:</div>
                    <input className="ant-input " onChange={@handleDeviceIDChange}></input>
                </div>
            </Modal>
        </div>

module.exports = DeviceCreater
