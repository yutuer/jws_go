CSVInput = require '../../common/csv_input'
antd        = require 'antd'
Api         = require '../api/api_ajax'
Modal       = antd.Modal
InputNumber = antd.InputNumber
RewardMaker = require '../../common/reward_maker'
TimeUtil    = require '../util/time'
React      = require 'react'

tokenTimeInit = "129600"
timeoutTimeInit = "10"
gzipSizeInit = "1024"
EndpointModal = React.createClass
    getInitialState: () ->
        {
            endpoint_server : @props.endpoint_server
            endpoint_chat   : @props.endpoint_chat
            data_ver        : @props.data_ver
            etcd_data_ver   : @props.etcd_data_ver
            bundle_ver      : @props.bundle_ver
            etcd_bundle_ver : @props.etcd_bundle_ver
            whitelistpwd    : @props.whitelistpwd
            biOption        : @props.biOption
            tokenTime       : @props.tokenTime ? tokenTimeInit
            timeoutTime     : @props.timeoutTime ? timeoutTimeInit
            prUrl           : @props.prUrl
            payUrls         : @props.payUrls
            gzip_size       : @props.gzip_size ? gzipSizeInit
            data_min        : @props.data_min
            bundle_min      : @props.bundle_min
            etcd_data_min   : @props.etcd_data_min
            etcd_bundle_min : @props.etcd_bundle_min
        }

    ShowModal : () ->
        @setState { visible: true }

    handleOk : () ->
        @setState { loading: true }
        @setState { loading: false, visible: false }
        @send()

    handleCancel : () ->
        @setState { visible: false }

    SetData : (data) ->
        @setState {
            endpoint_server : data.server
            endpoint_chat   : data.chat
            data_ver        : data.ver
            etcd_data_ver   : data.ver_etcd
            bundle_ver      : data.bundle_ver
            etcd_bundle_ver : data.bundle_ver_etcd
            whitelistpwd    : data.whitelistpwd
            biOption        : data.biOption
            tokenTime       : data.tokenTime
            timeoutTime     : data.timeoutTime
            prUrl           : data.pr_url
            payUrls         : data.pay_urls
            gzip_size       : data.gzip_size
            data_min        : data.data_min
            bundle_min      : data.bundle_min
            etcd_data_min   : data.etcd_data_min
            etcd_bundle_min : data.etcd_bundle_min
        }

    send : () ->
        console.log @state
        console.log @props

        tokenTime = @state.tokenTime
        timeoutTime = @state.timeoutTime
        gzipSize = @state.gzip_size

        tokenTime = tokenTimeInit if tokenTime is ""
        timeoutTime = timeoutTimeInit if timeoutTime is ""
        gzip_size = gzipSizeInit if gzipSize is ""
        api = new Api()
        api.Typ("setEndpoint")
           .ServerID(@props.server_id)
           .AccountID(@props.id)
           .Key(@props.curr_key)
           .Params(
               @props.gid,
               @props.version,
               @state.endpoint_server,
               @state.endpoint_chat,
               @state.data_ver,
               @state.whitelistpwd,
               @state.biOption,
               tokenTime,
               timeoutTime,
               @state.bundle_ver,
               @state.etcd_bundle_ver,
               @state.prUrl,
               @state.payUrls,
               gzipSize,
               @state.data_min,
               @state.bundle_min,
           )
           .Do (result) =>
                console.log "on_loaded"
                console.log result
                @props.on_loaded()
                dv = @state.data_ver
                bv = @state.bundle_ver
                @setState {
                    etcd_data_ver: dv
                    etcd_bundle_ver: bv
                }

    handleServerEndpointChange : (event) -> @setState { endpoint_server : event.target.value }
    handleChatEndpointChange   : (event) -> @setState { endpoint_chat   : event.target.value }
    handleDataVerChange        : (event) -> @setState { data_ver        : event.target.value }
    handleBundleVerChange      : (event) -> @setState { bundle_ver      : event.target.value }
    handleWhitelistpwdChange   : (event) -> @setState { whitelistpwd    : event.target.value }
    handleBIOptionChange       : (event) -> @setState { biOption        : event.target.value }
    handleTokenTimeChange      : (event) -> @setState { tokenTime       : event.target.value }
    handleTimeoutTimeChange    : (event) -> @setState { timeoutTime     : event.target.value }
    handlePrUrlChange          : (event) -> @setState { prUrl           : event.target.value }
    handlePayUrlsChange        : (v, str, is_right) -> @setState { payUrls : event.target.value}
    handleGZIPSizeChange       : (event) -> @setState { gzip_size       : event.toString() }
    handleDataMinChange        : (event) -> @setState { data_min        : event.target.value }
    handleBundleMinChange      : (event) -> @setState { bundle_min      : event.target.value }

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
                    <div>服务器Endpoint:</div>
                    <input className="ant-input" value = {@state.endpoint_server} onChange={@handleServerEndpointChange}></input>
                </div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>聊天ChatEndpoint:</div>
                    <input className="ant-input" value = {@state.endpoint_chat} onChange={@handleChatEndpointChange}></input>
                </div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>Data Version:</div>
                    <input className="ant-input" value = {@state.data_ver} onChange={@handleDataVerChange}></input>
                    <div>Data Version by Etcd:{@state.etcd_data_ver}</div>
                </div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>Data Min:</div>
                    <input className="ant-input" value = {@state.data_min} onChange={@handleDataMinChange}></input>
                    <div>Data Min by Etcd:{@state.etcd_data_min}</div>
                </div>

                <div className="row-flex row-flex-middle row-flex-start">
                    <div>Bundle Version:</div>
                    <input className="ant-input" value = {@state.bundle_ver} onChange={@handleBundleVerChange}></input>
                    <div>Bundle Version by Etcd:{@state.etcd_bundle_ver}</div>
                </div>

                <div className="row-flex row-flex-middle row-flex-start">
                    <div>Bundle Min:</div>
                    <input className="ant-input" value = {@state.bundle_min} onChange={@handleBundleMinChange}></input>
                    <div>Bundle Min by Etcd:{@state.etcd_bundle_min}</div>
                </div>

                <div className="row-flex row-flex-middle row-flex-start">
                    <div>whitelistpwd:</div>
                    <input className="ant-input" value = {@state.whitelistpwd} onChange={@handleWhitelistpwdChange}></input>
                </div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>大数据开关:</div>
                    <input className="ant-input" value = {@state.biOption} onChange={@handleBIOptionChange}></input>
                </div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>TokenTime(缺省129600):</div>
                    <input className="ant-input" value = {@state.tokenTime} onChange={@handleTokenTimeChange}></input>
                </div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>TimeoutTime(缺省10):</div>
                    <input className="ant-input" value = {@state.timeoutTime} onChange={@handleTimeoutTimeChange}></input>
                </div>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>PrUrl</div>
                    <input className="ant-input" value = {@state.prUrl} onChange={@handlePrUrlChange}></input>
                </div>
                <CSVInput {...@props} title="支付URL" value={@state.payUrls} ref = "payurl" can_cb = {@handlePayUrlsChange}/>
                <div className="row-flex row-flex-middle row-flex-start">
                    <div>GZIP Size(单位byte, 缺省1024):</div>
                    <InputNumber  min={0} value = {@state.gzip_size} onChange={@handleGZIPSizeChange}></InputNumber>
                </div>
            </Modal>


MkEndpointModal = React.createClass
    getInitialState: () -> {}
    componentDidMount: () -> @Refersh(0)

    Refersh: (time_wait) ->
        @setState {
            is_loading : true
        }

        if time_wait?
            setTimeout @Refersh, time_wait*1000
            return

        api = new Api()
        api.Typ("getEndpoint")
           .ServerID(@props.server_id)
           .AccountID(@props.id)
           .Key(@props.curr_key)
           .Params(
               @props.gid,
               @props.version
           )
           .Do (result, json, status) =>
                console.log "getEndpoint"
                console.log result
                console.log json
                if status is 200
                    @refs.modal.SetData json if json? and @refs.modal?
                else
                    @setState {
                        cannot : true
                    }

    onclickshow: () ->
        @Refersh(0)
        @refs.modal.ShowModal()

    render:() ->
        <div>
            <EndpointModal
                {...@props}
                id="null"
                ref="modal"
            />
            <button className="ant-btn ant-btn-primary" onClick={@onclickshow}>
                {@props.modal_name}
            </button>
        </div>

module.exports = MkEndpointModal
