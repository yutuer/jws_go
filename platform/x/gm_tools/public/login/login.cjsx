API = require '../tools/api/api_ajax'
MD5 = require 'md5'
React   = require 'react'
$       = require 'jquery'
antd    = require 'antd'
message = antd.message

success = ()    -> message.success('登陆成功');
error   = (msg) -> message.error('登陆失败:' + msg);

LoginApp = React.createClass
    getInitialState: () ->
        return {
            name   : ""
            passwd : ""
        }

    handleNameChange: (event) ->
        date =  event.target.value
        console.log date
        @setState {
            name : date
        }


    handlePassChange: (event) ->
        date =  event.target.value
        console.log date
        @setState {
            passwd : date
        }

    OnLogin: () ->
        self   = @
        name   = self.state.name
        passwd = self.state.passwd

        return if name == "" or passwd == ""

        $.ajax {
            url : "../api/v1/login/" + name,
            dataType: "json",
            type: "POST",
            data: JSON.stringify({"n":self.mkPw(passwd)}),
            contentType: "application/json; charset=utf-8",
            complete: (xhr, status) ->
                rc = {
                    status: xhr.status
                    json: xhr.responseJSON
                }
                console.log xhr
                console.log status
                if xhr.status is 200
                    success()
                    self.props.onChange(xhr.responseText)
                else
                    error(xhr.responseText)
        }

    mkPw : (passwd) ->
        return MD5(MD5(passwd) + "ta1hehud0ng")

    render:() ->
        <div>
            <hr />
                <div>姓名:</div>
                <input
                    className="ant-input"
                    value = {@state.priority}
                    onChange={@handleNameChange}></input>
                <div>密码:</div>
                <input
                    className="ant-input"
                    type="password"
                    value = {@state.priority}
                    onChange={@handlePassChange}></input>
                <button
                    key="submit"
                    className="ant-btn ant-btn-primary"
                    onClick={@OnLogin}>
                        提 交
                </button>
            <hr />
        </div>

module.exports = LoginApp
