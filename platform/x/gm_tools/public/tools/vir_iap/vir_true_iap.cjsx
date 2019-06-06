antd = require 'antd'
React      = require 'react'
Modal       = antd.Modal
InputNumber = antd.InputNumber

Api     = require '../api/api_ajax'

Select          = antd.Select

VirTrueIAP = React.createClass {
    getInitialState: () ->
        ss_init = {}
        ss_init["0"] = "Andriod"
        ss_init["1"] = "IOS"
        return {
          loading: false,
          visible: false,
          hc_add    : 1,
          iap       : 0,
          oper_showstate : "1"
          shard_show_state_const : ss_init  
        }

    showModal : () ->
        @setState { visible: true }

    handleOk : () ->
        @setState { loading: true }
        @setState { loading: false, visible: false }
        @sendtrueiap()

    handleCancel : () ->
        @setState { visible: false }

    sendtrueiap : () ->
        if @state.hc_add <= 0
            @props.onSend()
            return

        api = new Api()
        api.Typ("virtualtrueIAP")
           .ServerID(@props.server_id)
           .AccountID(@props.account_id)
           .Key(@props.curr_key)
           .Params(@state.hc_add,@state.iap,@state.oper_showstate)
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
    
    onSelectChange : (value) -> 
        console.log value
        @setState {
            oper_showstate : value
    }

    select_op : () ->
        res = []
        for k,v of @state.shard_show_state_const
            res.push(<Option value={k}>{v}</Option>)
        return res

    render:() ->
        <div>
            <antd.Button className="ant-btn ant-btn-primary" onClick={@showModal}>
                {@props.modal_name}
            </antd.Button>
            <Modal
                visible={@state.visible}
                title="发送真实模拟充值邮件" onOk={@handleOk} onCancel={@handleCancel}
                width=700

                footer={[
                    <antd.Button key="back" className="ant-btn" onClick={@handleCancel}>返 回</antd.Button>,
                    <antd.Button key="submit" className="ant-btn ant-btn-primary" onClick={@handleOk}>
                        提 交
                        <i className={'anticon anticon-loading'+(@state.loading?'':'hide')}></i>
                    </antd.Button>
                ]} >
                <p>已经发送的IAP充值邮件可以在单人邮件中查询</p>
                <p>充值ID:(通过IAP表中的IAPMAIN sheet 中的第一列ID，)</p>
                <InputNumber
                    min={1}
                    max={1000000}
                    defaultValue={1}
                    onChange={@handleRewardChange}
                    style={{width:70}} />

                <p>价钱(通过IAP表中的IAPBASE sheet 中的人民币价格一列，钱数要与所充值的物品对应，否则无效):</p>
                <InputNumber
                    defaultValue={0}
                    onChange={@handleIAPChange}
                    style={{width:70}} />

                <p>IOS or Android</p>
                <Select defaultValue={@state.oper_showstate} style={{ width: 120 }} onChange={@onSelectChange}>
                        {@select_op()}
                </Select>
            </Modal>
        </div>
}

module.exports = VirTrueIAP
