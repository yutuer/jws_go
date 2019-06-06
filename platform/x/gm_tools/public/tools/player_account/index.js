// Generated by CoffeeScript 1.12.7
(function() {
  var AccountKeyInput, Api, App, Button, ButtonToolbar, Col, DeviceEditor, DropdownButton, Grid, Input, ListGroup, ListGroupItem, MenuItem, Modal, ModalTrigger, NamePassEditor, Nav, NavItem, Navbar, React, ReactBootstrap, Row, Table, antd;

  antd = require('antd');

  Api = require('../api/api_ajax');

  AccountKeyInput = require('../../common/account_input');

  React = require('react');

  ReactBootstrap = require('react-bootstrap');

  ButtonToolbar = ReactBootstrap.ButtonToolbar;

  Button = ReactBootstrap.Button;

  MenuItem = ReactBootstrap.MenuItem;

  DropdownButton = ReactBootstrap.DropdownButton;

  Table = ReactBootstrap.Table;

  ModalTrigger = ReactBootstrap.ModalTrigger;

  Modal = ReactBootstrap.Modal;

  Navbar = ReactBootstrap.Navbar;

  Nav = ReactBootstrap.Nav;

  NavItem = ReactBootstrap.NavItem;

  Input = ReactBootstrap.Input;

  ListGroup = ReactBootstrap.ListGroup;

  ListGroupItem = ReactBootstrap.ListGroupItem;

  Grid = ReactBootstrap.Grid;

  Row = ReactBootstrap.Row;

  Col = ReactBootstrap.Col;

  NamePassEditor = require('./name_editor');

  DeviceEditor = require('./device_editor');

  App = React.createClass({displayName: "App",
    getInitialState: function() {
      return {
        select_server: this.props.curr_server,
        player_to_send: "",
        name_to_find: "",
        device_to_find: "",
        account_data: {}
      };
    },
    handleServerChange: function(data) {},
    handleNameChange: function() {
      return this.setState({
        name_to_find: this.refs.name_input.getValue()
      });
    },
    handleDeviceChange: function() {
      return this.setState({
        device_to_find: this.refs.device_input.getValue()
      });
    },
    getServerName: function() {
      if (this.state.select_server != null) {
        return this.state.select_server.name;
      } else {
        return "";
      }
    },
    queryByDeciveID: function() {
      var api, id;
      id = this.refs.device_input.getValue();
      api = new Api();
      console.log(this.state.send);
      return api.Typ("getAccountByDeviceID").ServerID("SID").AccountID("AID").Key(this.props.curr_key).Params(id).Do((function(_this) {
        return function(result) {
          console.log("on queryByDeciveID");
          console.log(result);
          return _this.setState({
            account_data: JSON.parse(result)
          });
        };
      })(this));
    },
    queryByName: function() {
      var api, name;
      name = this.refs.name_input.getValue();
      api = new Api();
      console.log(this.state.send);
      return api.Typ("getAccountByName").ServerID("SID").AccountID("AID").Key(this.props.curr_key).Params(name).Do((function(_this) {
        return function(result) {
          console.log("on queryByName");
          console.log(result);
          return _this.setState({
            account_data: JSON.parse(result)
          });
        };
      })(this));
    },
    getAccountRes: function() {
      if ((this.state.account_data != null) && (this.state.account_data.user_id != null)) {
        return this.state.account_data.user_id;
      }
      return "null";
    },
    render: function() {
      return React.createElement("div", null, React.createElement("div", null, React.createElement(Input, {
        "type": 'text',
        "value": this.state.name_to_find,
        "placeholder": '请输入玩家注册名...',
        "help": '通过用户名查找UID, UID在下面显示',
        "bsStyle": 'success',
        "hasFeedback": true,
        "ref": 'name_input',
        "onChange": this.handleNameChange
      }), React.createElement(Button, {
        "bsStyle": 'primary',
        "onClick": this.queryByName
      }, "查找注册名")), React.createElement("div", null, React.createElement(Input, {
        "type": 'text',
        "value": this.state.device_to_find,
        "placeholder": '请输入DeviceID...',
        "help": '通过DeviceID查找UID, UID在下面显示',
        "bsStyle": 'success',
        "hasFeedback": true,
        "ref": 'device_input',
        "onChange": this.handleDeviceChange
      }), React.createElement(Button, {
        "bsStyle": 'primary',
        "onClick": this.queryByDeciveID
      }, "查找DeviceID")), React.createElement("div", null, "用户UserID : " + this.getAccountRes()), React.createElement(NamePassEditor, Object.assign({}, this.props)), React.createElement(DeviceEditor, Object.assign({}, this.props)));
    }
  });

  module.exports = App;

}).call(this);
