antd   = require 'antd'
React  = require 'react'
Modal  = antd.Modal
Api    = require '../api/api_ajax'

NamePassCreater = React.createClass
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
        api.Typ("createNamePassForUID")
           .ServerID("SID")
           .AccountID(@state.user_id)
           .Key(@props.curr_key)
           .Params(@state.name)
           .Do (result) =>
                console.log "on createNamePassForUID"
                console.log result

    handleUserIDChange : (event) -> @setState { user_id : event.target.value }
    handleNameChange   : (event) -> @setState { name    : event.target.value }

    render:() ->
        <div>
            <button className="ant-btn ant-btn-primary" onClick={@showModal}>
                绑定UId到用户名
            </button>
            <Modal ref="modal"
                visible={@state.visible}
                title="绑定UId到用户名" onOk={@handleOk} onCancel={@handleCancel}
                width=700
                footer={[
                    <button key="back" className="ant-btn" onClick={@handleCancel}>返 回</button>,
                    <button key="submit" className="ant-btn ant-btn-primary" onClick={@handleOk}>
                        提 交
                        <i className={'anticon anticon-loading'+(@state.loading?'':'hide')}></i>
                    </button>
                ]}>
                <div>注意:绑绑定UId到用户名之前要查询用户名是否存在, 如果用户名存在, 会覆盖原有信息. 这里会将密码重置为1, 建议不要覆盖原有用户名</div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>UserID(xx-xx-xx):</div>
                    <input className="ant-input " onChange={@handleUserIDChange}></input>
                </div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>Name:</div>
                    <input className="ant-input " onChange={@handleNameChange}></input>
                </div>
            </Modal>
        </div>

module.exports = NamePassCreater
