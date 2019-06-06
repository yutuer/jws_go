(function() {
    var App, Api, React, antd, Modal, Button, Checkbox, CheckboxGroup, boot, Row, Col, Icon, Table;
    antd = require('antd');

    React = require('react');

    boot = require('react-bootstrap');

    Api = require('../api/api_ajax');

    Modal = antd.Modal;

    Button = antd.Button;

    Icon = antd.Icon;

    Checkbox = boot.Checkbox;

    CheckboxGroup = boot.FormGroup;

    Row = boot.Row;

    Col = boot.Col;

    Table = antd.Table;

    App = React.createClass({displayName: "App",
        getInitialState:function () {
              return {
                  createServerId:"",
                  createAccountId:"",
                  grant_id:"",
                  modal1Visible: false,
                  modal2Visible: false,
                  modal3Visible: false,
              }
        },
        setModal1Visible:function(modal1Visible) {
            this.setState({ modal1Visible });
        },
        setModal2Visible:function(modal2Visible) {
            this.setState({ modal2Visible });
        },
        setModal3Visible:function(modal3Visible) {
            this.setState({ modal3Visible });
        },
        sendCreateServerID:function (event) {
            this.setState({
                createServerId:event.target.value
            })
        },
        sendCreateAccountID:function (event) {
            this.setState({
                createAccountId:event.target.value
            })
        },
        sendCreateAccount:function () {
            this.setState({
                modal1Visible:false
            })
            var api;
            api = new Api();
            return api.Typ("addUser").ServerID("").AccountID("").Key(this.props.curr_key).Params(this.state.createServerId,"admin",this.state.createAccountId).Do((function(_this) {
                return function(result) {
                    console.log(result);
                };
            })(this));
        },
        getUserName:function (e) {
            this.setState({
                UserName:e.target.value
            })
        },
        getGrantValue:function(){
            var str = document.getElementsByName("grant_value");
            var grantArray = str.length;
            var grantValue = "";
            for (var i = 0; i < grantArray ; i++){
                if (str[i].checked == true){
                    grantValue+=str[i].value + "+";
                }
            }
            grantValue = grantValue.substring(0,grantValue.length-1);
            this.setState({
                grant_id:grantValue,
            })
        },
        getGrant:function () {
            var api;
            api = new Api();
            return api.Typ("getGrant").ServerID("").AccountID("").Key(this.props.curr_key).Params(this.state.UserName).Do((function(_this) {
                return function(result,json) {
                    var arr ="";
                    for(var j = 0; j < json.length; j++) {
                        if (json[j] == "1" ) {
                            arr += j + "+";
                        }
                    }
                    arr = arr.substring(0,arr.length-1);
                    arr = arr.split("+");
                    var chec = document.getElementsByName("grant_value");
                    var grantArray = chec.length;
                    var graValArr = new Array();
                    var checString = "";
                    for (var i = 0; i < grantArray ; i++){
                            graValArr[i] = chec[i].value.split("+");
                            if(arr.indexOf(graValArr[i][0])!=-1){
                                chec[i].checked = true;
                                checString+=chec[i].value + "+"
                            }else{;
                                chec[i].checked = false;
                            }
                    }
                    checString = checString.substring(0,checString.length-1);
                    _this.setState({
                        grant_id:checString,
                    })
                };
            })(this));
        },
        setGrant:function () {
            var api;
            api = new Api();
            this.setState({
                modal2Visible:false,
            });
            return api.Typ("setGrant").ServerID("").AccountID("").Key(this.props.curr_key).Params(this.state.UserName,this.state.grant_id).Do((function() {
                return function(result) {
                    console.log(result);
                };
            })(this));
        },
        getPassWord:function () {
            var btn = document.getElementById("passwordBtn");
            var pass = document.getElementById("passwordInput");
            btn.onmousedown = function(){
                pass.type = "number"
            };
            btn.onmouseup = btn.onmouseout = function(){
                pass.type = "password"
            }
        },
        getOperation:function () {

        },
        render() {
            let groupStyle = {
              marginTop:-10
            };
            let modelStyle = {
              marginTop:30
            };
            return (
                <div>
                    <Button type = "primary" onClick = { () => this.setModal1Visible(true) }>创建账号</Button>
                    <Modal
                        title = "创建账号"
                        visible = {this.state.modal1Visible}
                        onOk = {() => this.sendCreateAccount()}
                        onCancel = {() => this.setModal1Visible(false)}
                    >
                        <p>账号:<input
                        onChange = { this.sendCreateServerID }
                        value = { this.state.createServerId }
                        >
                        </input></p>
                        <p>密码:<input type="password" id="passwordInput"
                        onChange = { this.sendCreateAccountID }
                        >
                        </input> <Button onClick = { this.getPassWord } id="passwordBtn" shape="circle"><Icon type="eye-o" /></Button></p>
                        <h5>注意:新注册的账号是空权限账号,如果想赋予权限请点击下面更改权限按钮</h5>
                    </Modal>
                    <br/>
                    <hr/>
                    <Button type = "primary" onClick = { () => this.setModal2Visible(true) }>更改权限</Button>
                    <Modal
                        title = "更改权限"
                        visible = { this.state.modal2Visible }
                        onOk = { () => this.setGrant() }
                        onCancel = {() => this.setModal2Visible(false)}
                    >
                        <h5>用户名 : <input className="example-input" type="text"
                                      onChange={this.getUserName}
                        value={ this.state.UserName }/>
                            <Button type = "primary" onClick = { this.getGrant } >查询权限 </Button>
                            <br/>
                            <p>Notice:赋予或者查询权限时,请先输入用户名,再进行操作,单人邮件与全服邮件的删除按钮是一个权限,批量邮件的权限是获取批量号与下载批量邮件</p>
                        </h5>
                        <div>
                        <CheckboxGroup style={ groupStyle } value = { this.state.grant_id } onChange = { this.getGrantValue }>
                                <Col style={ modelStyle } xs={6} md={6}><Checkbox disabled="false">权限管理(未开启)</Checkbox></Col>
                                <Col style={ modelStyle } xs={6} md={6}><Checkbox name="grant_value" value="20+64">昵称\ID查询</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="38">删除邮件权限</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="59+60">批量邮件</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="3+4+5+56">虚拟充值</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="10+11+12+13+14+53">跑马灯公告</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="6+7+8+9+17+18+42+58">公告</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="15">封禁账号</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="19">禁言账号</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox disabled="false" value="">批量封禁(默认权限)</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="26+27+28+29">账号查询</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="23+24+25">账号转移</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="21+22">渠道链接</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="30+31+32+45">支付查询</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox disabled="false" value="">Shard管理(暂时关闭权限赋予)</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox disabled="false" value="">Shard状态标签(暂时关闭权限赋予)</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="43+44">活动开关</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="46+47+48+49+50">数据版本</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="51+52">删除排行榜</Checkbox></Col>
                                <Col xs={6} md={6}><Checkbox name="grant_value" value="55">设备转移</Checkbox></Col>
                        </CheckboxGroup>
                        </div>
                    </Modal>
                    <br/>
                    <hr/>
                    <Button disabled  type = "primary" onClick = { () => this.setModal3Visible(true) } >查看操作日志(暂时不可用) </Button>
                    <Modal
                        title = "查看操作日志"
                        visible = { this.state.modal3Visible }
                        onOk = { () => this.setModal3Visible(false) }
                        onCancel = {() => this.setModal3Visible(false)}
                    >
                        <h5>用户名 : <input className="example-input" type="text"
                                         onChange={this.getUserName}
                                         value={ this.state.UserName }/>
                        <Button disabled type = "primary" onClick={ () => this.getOperation } >查询操作日志</Button></h5>
                    </Modal>
                </div>
            )
        }
    });

    module.exports = App;

}).call(this);
