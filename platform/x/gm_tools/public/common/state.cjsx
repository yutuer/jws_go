React      = require 'react'
# 正常 -> 加载中 -> 加载完毕回调 -> 正常
class ReactLoadingState
    constructor: ( @component )->

    # 进入Loading状态
    EnterLoading: ( typ, leave_callback ) ->
        console.log "EnterLoading ", typ
        loadstat__ = @component.state.loadstat__
        loadstat__ = {} if not loadstat__?
        loadstat__[typ] = [1, leave_callback]
        @component.setState {
            loadstat__ : loadstat__
        }
        return

    # 离开Loading状态
    LeaveLoading: ( typ ) ->
        console.log "LeaveLoading ", typ
        loadstat__ = @component.state.loadstat__
        cb = loadstat__[typ][1]
        cb() if cb?
        loadstat__[typ] = undefined
        @component.setState {
            loadstat__ : loadstat__
        }

    isLoading : ( typ ) ->
        return false if not @component.state.loadstat__?
        state = @component.state.loadstat__[typ]
        if state?
            return state[0] is 1
        else
            return false

    GetBtnStat : ( typ ) ->
        if @isLoading typ
            return 'ant-btn-loading'
        else
            return ''




module.exports = {
    ReactLoadingState : ReactLoadingState
}
