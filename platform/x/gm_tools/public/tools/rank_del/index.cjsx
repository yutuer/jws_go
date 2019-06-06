antd    = require 'antd'
Api     = require '../api/api_ajax'
React   = require 'react'
AccountKeyInput = require '../../common/account_input'
ReactBootstrap = require 'react-bootstrap'
Button         = ReactBootstrap.Button
Row            = antd.Row
Col            = antd.Col
Form            = antd.Form
FormItem        = antd.Form.Item
Input           = antd.Input
Button          = antd.Button

App = React.createClass
    getInitialState: () ->
        return {
            select_server : @props.curr_server
            player_to_send : ""
            rank_name: ""
            param: ""
        }

    handleUserChange: (data) ->
        console.log data
        if data is ""
            data = "请输入玩家Id"
        @setState {
            player_to_send : data
        }

    handleServerChange: (data) ->
        console.log data
        @setState {
            select_server : data
        }
    validationState: (data) ->
        if not @state.rank_name
              return 'error'
        return 'success'

    rank_delete: () ->
        console.log @state.select_server
        api = new Api()
        api.Typ("delRank")
         .ServerID(@state.select_server.serverName)
         .AccountID(@state.player_to_send)
         .Key(@props.curr_key)
         .Params(@state.rank_name, @state.param)
         .Do (result) =>
            console.log result

    rank_reload : () ->
        api = new Api()
        api.Typ("reloadRank")
         .ServerID(@state.select_server.serverName)
         .AccountID(@state.player_to_send)
         .Key(@props.curr_key)
         .ParamArray()
         .Do (result) =>
            console.log result
    handleChange : (event) ->
         @setState {
            rank_name : event.target.value
         }
    handleParamChange : (event) ->
         @setState {
            param: event.target.value
         }
    render:() ->
        <div>
            <p>说明：</p>
            <p>【删除排行榜】：会删除数据库中对应的角色id，并重新加载排行榜，非专业人士，请勿操作！</p>
             <Form inline onSubmit={@handleSubmit_data}>
                            <FormItem label="排行榜:">
                              <ReactBootstrap.Input type='text' className = "col-offset-1" placeholder="" style={{width:243}} bsStyle={@validationState()} onChange={@handleChange}'
                              />
                            </FormItem>
                            <FormItem label="参数:" className = "col-offset-4">
                              <ReactBootstrap.Input type='text' className = "col-offset-1" placeholder="" style={{width:155}} onChange={@handleParamChange}}
                              />
                            </FormItem>
                        </Form>
            <AccountKeyInput {...@props} ref = "accountin" can_cb = {@handleUserChange}/>
           
            <Row>
              <Col span="4" offset="18" >
                <Button
                    bsStyle='danger'
                    onClick= @rank_delete
                    disabled={ @state.player_to_send == "" || @state.rank_name == ""}
                >
                    从排行榜上删除
                </Button>
              </Col>
            </Row>
        </div>


module.exports = App