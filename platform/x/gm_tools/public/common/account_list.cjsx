antd = require 'antd'
boot = require 'react-bootstrap'
React     = require 'react'
FileInput = require 'react-file-input'

InputNumber = antd.InputNumber
Row    = antd.Row
Table  = boot.Table
Upload = antd.Upload
Icon   = antd.Icon
Button = antd.Button

MaxAccountCount = 50

AccountList = React.createClass
    getInitialState: () ->
        return {
            accounts : []
            curr_id  : null
        }

    onInputActChange: (event) ->
        @setState {
            curr_id : event.target.value
        }

    addAccount : () ->
        id = @state.curr_id
        r = @state.accounts
        r.push id
        @setState {
            accounts : r
        }
        @props.onChange r if @props.onChange?

    cannotAddAccount : () ->

    getAddButton : () ->
        if @state.curr_id? and @state.curr_id isnt "" and @state.accounts.length < MaxAccountCount
            return <button className="ant-btn ant-btn-primary"  onClick={@addAccount}>增加</button>
        else
            return <button className="ant-btn disable"  onClick={@cannotAddAccount}>增加</button>

    del : (k) ->
        return () =>
            nr = []
            for ki, v of @state.accounts
                nr.push v if ki isnt k
            @setState {
                accounts : nr
            }
            @props.onChange nr if @props.onChange?


    getAllInfo : (data) ->
        if not data?
          return <div>UnKnown Info</div>

        re = []
        for k, v of data
          re.push <tr key={k} >
                  <td>{v}</td>
                  <td>
                        <boot.Button bsStyle='danger'
                                      disabled = {false}
                                      onClick={ @del(k)} >
                                      删除
                        </boot.Button>
                  </td>
              </tr>
        return re

    handleCSVFileChange : ( info ) ->
        if info.file.status != 'uploading'
            reader = new FileReader()
            self = @
            reader.onload = () ->
                console.log(@result)
                res = @result.split('\n')
                console.log('Selected file:', res)
                r = self.state.accounts
                for k,v of res
                    r.push v if r.length < 50 and v isnt ""
                console.log('Selected file:', r)
                self.setState {
                    accounts : r
                }
                self.props.onChange r if self.props.onChange?
                return
            reader.readAsText(info.file.originFileObj)


    getAccountList : () ->
        <Table striped bordered condensed hover>
            <thead>
                <tr>
                    <th>AccountID</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody>
                { @getAllInfo(@state.accounts) }
            </tbody>
        </Table>

    render:() ->
        <div>
            <div className="row-flex row-flex-middle row-flex-start">
                <div>AccountID:</div>
                <input
                        span="8"
                        className="ant-input"
                        type="text"
                        id="ItemID"
                        placeholder="请输入AccountID"
                        onChange={@onInputActChange} />
                {@getAddButton()}
                <Upload onChange={@handleCSVFileChange} >
                    <Button type="ghost">
                        <Icon type="upload" /> 点击加载文件
                    </Button>
                </Upload>
                <div>加载文件是txt文件,每一行(\n分隔)是一个AccountID,类似于"0:10:XXXX"这样的</div>

            </div>
            <hr />
            {@getAccountList()}
        </div>

module.exports = AccountList
