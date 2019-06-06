antd = require 'antd'
boot = require 'react-bootstrap'
React      = require 'react'

InputNumber = antd.InputNumber
Table = boot.Table

RewardMaker = React.createClass
    getInitialState: () ->
        return {
            rewards : []
            curr_id : null
            curr_num : 1
        }

    onInputNumberChange: (num) ->
        @setState {
            curr_num : num
        }

    onInputItemIDChange: (event) ->
        @setState {
            curr_id : event.target.value
        }

    addReward : () ->
        id = @state.curr_id
        count = @state.curr_num
        r = @state.rewards
        r.push {
            Id : id,
            Count : count
        }
        @setState {
            rewards : r
        }
        @props.onChange r if @props.onChange?

    cannotAddReward : () ->

    getAddButton : () ->
        if @state.curr_id? and @state.curr_id isnt "" and @state.rewards.length < 5
            return <button className="ant-btn ant-btn-primary"  onClick={@addReward}>增加</button>
        else
            return <button className="ant-btn disable"  onClick={@cannotAddReward}>增加</button>

    del : (k) ->
        return () =>
            nr = []
            for ki, v of @state.rewards
                nr.push v if ki isnt k
            @setState {
                rewards : nr
            }
            @props.onChange nr if @props.onChange?


    getAllInfo : (data) ->
        if not data?
          return <div>UnKnown Info</div>

        re = []
        for k, v of data
          re.push <tr key={k} >
                  <td>{v.Id}</td>
                  <td>{v.Count}</td>
                  <td className='row-flex row-flex-middle row-flex-start'>
                        <boot.Button bsStyle='danger'
                                      disabled = {false}
                                      onClick={ @del(k)} >
                                      删除
                        </boot.Button>
                  </td>
              </tr>
        return re

    getRewardList : () ->
        <Table striped bordered condensed hover>
            <thead>
                <tr>
                    <th>ItemID</th>
                    <th>数量</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody>
                { @getAllInfo(@state.rewards) }
            </tbody>
      </Table>

    render:() ->
        <div>
            <div className="row-flex row-flex-middle row-flex-start">
                <div className="col-2">类型ID:</div>
                <input
                    className="ant-input col-12"
                    type="text"
                    id="ItemID"
                    placeholder="请输入类型ID"
                    onChange={@onInputItemIDChange} />
                <div className="col-2" />
                <div className="col-2">数量:</div>
                <InputNumber
                    min={1}
                    max={1000000}
                    defaultValue={1}
                    onChange={@onInputNumberChange}
                    style={{width:70}}/>
                {@getAddButton()}
            </div>
            <hr />
            {@getRewardList()}
        </div>

module.exports = RewardMaker
