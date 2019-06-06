antd = require 'antd'
React      = require 'react'
Modal       = antd.Modal
InputNumber = antd.InputNumber

Api     = require '../api/api_ajax'


NewSendIAPModal = React.createClass {
    getInitialState: () ->
        return {
          loading: false,
          visible: false,
          hc_add    : 1,
          iap       : 0,
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
        if @state.hc_add <= 0
            @props.onSend()
            return

        api = new Api()
        api.Typ("virtualIAP")
           .ServerID(@props.server_id)
           .AccountID(@props.account_id)
           .Key(@props.curr_key)
           .Params(@state.hc_add)
           .Do (result) =>
                @setState {
                    mails : JSON.parse(result)
                    is_loading : false
                }
                console.log "onSend"
                @props.onSend()

    handleRewardChange : (count) ->
        @setState {
            hc_add : parseInt(count, 10)
        }
        console.log "handleRewardChange"
        console.log count

    handleIAPChange : (count) ->
        @setState {
            iap : parseInt(count, 10)
        }
        console.log "handleIAP"
        console.log count

    render:() ->
        <div>
            <antd.Button className="ant-btn ant-btn-primary" onClick={@showModal}>
                {@props.modal_name}
            </antd.Button>
            <Modal
                visible={@state.visible}
                title="发送模拟充值邮件" onOk={@handleOk} onCancel={@handleCancel}
                width=700

                footer={[
                    <antd.Button key="back" className="ant-btn" onClick={@handleCancel}>返 回</antd.Button>,
                    <antd.Button key="submit" className="ant-btn ant-btn-primary" onClick={@handleOk}>
                        提 交
                        <i className={'anticon anticon-loading'+(@state.loading?'':'hide')}></i>
                    </antd.Button>
                ]} >

                <p>奖励:</p>
                <InputNumber
                    min={1}
                    max={1000000}
                    defaultValue={1}
                    onChange={@handleRewardChange}
                    style={{width:70}} />

            </Modal>
        </div>
}

module.exports = NewSendIAPModal
