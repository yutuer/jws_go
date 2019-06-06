// Generated by CoffeeScript 1.12.7
(function() {
  var Api, App, Button, CSVInput, Input, React, Table, antd, boot;

  antd = require('antd');

  boot = require('react-bootstrap');

  CSVInput = require('../../common/csv_input');

  Api = require('../api/api_ajax');

  React = require('react');

  Input = boot.Input;

  Table = boot.Table;

  Button = antd.Button;

  App = React.createClass({displayName: "App",
    getInitialState: function() {
      return {
        select_server: this.props.curr_server,
        send: [],
        res: [],
        content: ""
      };
    },
    isServerRight: function() {
      return (this.state.select_server != null) && (this.state.select_server.serverName != null);
    },
    getLoadingState: function() {
      if (!this.isServerRight()) {
        return "disabled";
      }
      return '';
    },
    handleServerChange: function(data) {
      return this.setState({
        select_server: data
      });
    },
    handleChange: function(v, str, is_right) {
      var nick_names;
      nick_names = v.map(function(a) {
        return a[0];
      });
      return this.setState({
        send: nick_names,
        content: str
      });
    },
    getServerName: function() {
      if (this.state.select_server != null) {
        return this.state.select_server.name;
      } else {
        return "";
      }
    },
    getNick: function() {
      var api;
      api = new Api();
      return api.Typ("getInfoByNickName").ServerID(this.state.select_server.serverName).AccountID("AID").Key(this.props.curr_key).ParamArray(this.state.send).Do((function(_this) {
        return function(result) {
          console.log("onSend");
          return _this.setState({
            res: JSON.parse(result)
          });
        };
      })(this));
    },
    getAcid: function () {
      var api;
      api = new Api();
      return api.Typ("getNickNameFromRedisByACID").ServerID(this.state.select_server.serverName).AccountID("").Key(this.props.curr_key).ParamArray(this.state.send).Do((function(_this) {
          return function(result) {
              console.log("onSend");
              return _this.setState({
                  res: JSON.parse(result)
              });
          };
      })(this));
    },
    getRes: function() {
      return React.createElement(Table, Object.assign({}, this.props, {
        "striped": true,
        "bordered": true,
        "condensed": true,
        "hover": true
      }), React.createElement("thead", null, React.createElement("tr", null, React.createElement("th", null, "昵称"), React.createElement("th", null, "ID"))), React.createElement("tbody", null, this.state.res.map(function(v) {
          return React.createElement("tr", {
          "key": v.name
        }, React.createElement("td", null, " ", v.name, " "), React.createElement("td", null, " ", v.pid, "  "));
      })));
    },
    render: function() {
      return React.createElement("div", null, React.createElement(CSVInput, Object.assign({}, this.props, {
        "title": "请输入昵称/ID",
        "value": this.state.content,
        "ref": "accountin",
        "can_cb": this.handleChange
      })), React.createElement(Button, {
        "className": 'ant-btn ant-btn-primary ' + this.getLoadingState(),
        "onClick": this.getNick
      },"查询Acid"),
          React.createElement(Button, {
          "className": 'ant-btn ant-btn-primary ' + this.getLoadingState(),
          "onClick": this.getAcid
      },"查询昵称"),
          this.getRes());
    }
  });

  module.exports = App;

}).call(this);
