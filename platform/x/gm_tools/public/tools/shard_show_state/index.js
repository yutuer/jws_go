// Generated by CoffeeScript 1.12.7
(function() {
  var Api, App, Button, Checkbox, CheckboxGroup, Col, Form, FormItem, Grid, Option, React, ReactBootstrap, Row, Select, Table, antd, tbody, td, tr;

  antd = require('antd');

  Api = require('../api/api_ajax');

  React = require('react');

  ReactBootstrap = require('react-bootstrap');

  Form = antd.Form;

  FormItem = Form.Item;

  Select = antd.Select;

  Option = Select.Option;

  Grid = ReactBootstrap.Grid;

  Row = ReactBootstrap.Row;

  Col = ReactBootstrap.Col;

  Button = antd.Button;

  Table = ReactBootstrap.Table;

  tbody = ReactBootstrap.tbody;

  tr = ReactBootstrap.tr;

  td = ReactBootstrap.td;

  Checkbox = antd.Checkbox;

  CheckboxGroup = Checkbox.Group;

  App = React.createClass({displayName: "App",
    getInitialState: function() {
      var k, r_ss_init, ss_init, teamAB_init, v;
      ss_init = {};
      ss_init["0"] = "新服";
      ss_init["1"] = "流畅";
      ss_init["2"] = "火爆";
      ss_init["3"] = "拥挤";
      ss_init["4"] = "维护";
      ss_init["5"] = "体验";
      r_ss_init = [];
      for (k in ss_init) {
        v = ss_init[k];
        r_ss_init[v] = k;
      }
      teamAB_init = {};
      teamAB_init["0"] = "A";
      teamAB_init["1"] = "B";
      teamAB_init["2"] = "AB";
      return {
        shard_show_state_const: ss_init,
        r_shard_show_state_const: r_ss_init,
        shard_show_teamAB: teamAB_init,
        gids: [],
        curTeamAB: [],
        select_gid: "",
        sid_ss: [],
        selected_sid: {},
        op_show_state: "0",
        op_show_teamAB: "0"
      };
    },
    componentDidMount: function() {
      var api;
      api = new Api();
      return api.Typ("getAllGids").ServerID("").AccountID("").Key(this.props.curr_key).ParamArray().Do((function(_this) {
        return function(result) {
          var def, res;
          console.log(result);
          res = JSON.parse(result);
          if (res.length > 0) {
            def = String(res[0]);
          }
          return _this.setState({
            gids: res,
            select_gid: def
          });
        };
      })(this));
    },
    onGidSelectChange: function(value) {
      var api;
      console.log(value);
      this.setState({
        select_gid: value
      });
      api = new Api();
      return api.Typ("getShardShowStateByGid").ServerID("").AccountID("").Key(this.props.curr_key).ParamArray([value]).Do((function(_this) {
        return function(result) {
          var res;
          console.log(result);
          res = JSON.parse(result);
          return _this.setState({
            sid_ss: res,
            selected_sid: {}
          });
        };
      })(this));
    },
    selectOption: function() {
      var gid, i, len, ref, res, str;
      res = [];
      ref = this.state.gids;
      for (i = 0, len = ref.length; i < len; i++) {
        gid = ref[i];
        str = String(gid);
        res.push(React.createElement(Option, {
          "value": str
        }, str));
      }
      return res;
    },
    selectSid: function(e) {
      var i, len, ss, value;
      ss = this.state.selected_sid;
      for (i = 0, len = e.length; i < len; i++) {
        value = e[i];
        ss[value] = value;
      }
      console.log(ss);
      return this.setState({
        selected_sid: ss
      });
    },
    genGroupCheckBox: function(ar) {
      var a, i, len, plainOptions;
      plainOptions = [];
      for (i = 0, len = ar.length; i < len; i++) {
        a = ar[i];
        plainOptions.push(this.decodeShowState(a));
      }
      console.log(plainOptions);
      return React.createElement("tr", null, React.createElement("td", null, React.createElement(CheckboxGroup, {
        "options": plainOptions,
        "onChange": this.selectSid
      })));
    },
    gencheckbox: function() {
      var column_count, count, i, len, ref, res, s, selected;
      column_count = 5;
      res = [];
      selected = [];
      count = 0;
      ref = this.state.sid_ss;
      for (i = 0, len = ref.length; i < len; i++) {
        s = ref[i];
        count = count + 1;
        if (count <= column_count) {
          selected.push(s);
        }
        if (count === column_count) {
          res.push(this.genGroupCheckBox(selected));
          selected = [];
          count = 0;
        }
      }
      if (selected.length > 0) {
        res.push(this.genGroupCheckBox(selected));
        console.log(res);
      }
      return res;
    },
    show_state_select_op: function() {
      var k, ref, res, v;
      res = [];
      ref = this.state.shard_show_state_const;
      for (k in ref) {
        v = ref[k];
        if (k === "4" || k === "5") {
          continue;
        }
        res.push(React.createElement(Option, {
          "value": k
        }, v));
      }
      return res;
    },
    show_teamAB_select_op: function() {
      var k, ref, res, v;
      res = [];
      ref = this.state.shard_show_teamAB;
      for (k in ref) {
        v = ref[k];
        res.push(React.createElement(Option, {
          "value": k
        }, v));
      }
      return res;
    },
    onChgShowState: function(value) {
      return this.setState({
        op_show_state: value
      });
    },
    onChgTeamAB: function(value) {
      return this.setState({
        op_show_teamAB: value
      });
    },
    onSetShowTeamAB: function() {
      var api, k, ref, ss_ar, v;
      ss_ar = [];
      ss_ar.push(this.state.select_gid);
      ss_ar.push(this.state.shard_show_teamAB[this.state.op_show_teamAB]);
      console.log(this.state.selected_sid);
      ref = this.state.selected_sid;
      for (k in ref) {
        v = ref[k];
        if (v != null) {
          ss_ar.push(this.encodeShowState(k));
        }
      }
      console.log(ss_ar);
      if (ss_ar.length > 0) {
        api = new Api();
        return api.Typ("setShardsShowTeamAB").ServerID("").AccountID("").Key(this.props.curr_key).ParamArray(ss_ar).Do((function(_this) {
          return function(result) {
            var res;
            console.log(result);
            res = JSON.parse(result);
            return _this.setState({
              sid_ss: res,
              selected_sid: {}
            });
          };
        })(this));
      }
    },
    onSetShowState: function() {
      var api, k, ref, ss_ar, v;
      ss_ar = [];
      ss_ar.push(this.state.select_gid);
      ss_ar.push(this.state.op_show_state);
      console.log(this.state.selected_sid);
      ref = this.state.selected_sid;
      for (k in ref) {
        v = ref[k];
        if (v != null) {
          ss_ar.push(this.encodeShowState(k));
        }
      }
      console.log(ss_ar);
      if (ss_ar.length > 0) {
        api = new Api();
        return api.Typ("setShardsShowState").ServerID("").AccountID("").Key(this.props.curr_key).ParamArray(ss_ar).Do((function(_this) {
          return function(result) {
            var res;
            console.log(result);
            res = JSON.parse(result);
            return _this.setState({
              sid_ss: res,
              selected_sid: {}
            });
          };
        })(this));
      }
    },
    encodeShowState: function(str) {
      var ss;
      ss = str.split(" ");
      return ss[0] + " " + this.state.r_shard_show_state_const[ss[1]] + ss[2];
    },
    decodeShowState: function(str) {
      var show_state, ss;
      ss = str.split(" ");
      show_state = this.state.shard_show_state_const[ss[1]];
      if (!(show_state != null)) {
        show_state = "";
      }
      return ss[0] + " " + show_state + ss[2];
    },
    render: function() {
      return React.createElement("div", null, React.createElement("p", null, "说明："), React.createElement("p", null, "体验和维护 这两个状态这个页面是没有权限修改的"), React.createElement("br", null), React.createElement(Form, {
        "horizontal": true
      }, React.createElement(FormItem, {
        "id": "select",
        "label": "Gid：",
        "labelCol": {
          span: 2
        },
        "wrapperCol": {
          span: 14
        }
      }, React.createElement(Select, {
        "id": "select",
        "size": "large",
        "style": {
          width: 120
        },
        "onChange": this.onGidSelectChange
      }, this.selectOption()))), React.createElement("div", {
        "style": {
          marginBottom: 24
        }
      }), React.createElement("div", null, React.createElement(Grid, null, React.createElement(Row, null, React.createElement(Col, {
        "xs": 12,
        "md": 8
      }, React.createElement(Table, {
        "responsive": true
      }, React.createElement("tbody", null, this.gencheckbox()))), React.createElement(Col, {
        "xs": 6,
        "md": 4
      }, React.createElement(Select, {
        "id": "select",
        "size": "large",
        "defaultValue": this.state.op_show_state,
        "style": {
          width: 120
        },
        "onChange": this.onChgShowState
      }, this.show_state_select_op()), React.createElement(Select, {
        "id": "select",
        "size": "large",
        "defaultValue": this.state.op_show_state,
        "style": {
          width: 120
        },
        "onChange": this.onChgTeamAB
      }, this.show_teamAB_select_op()), React.createElement("br", null), React.createElement(Button, {
        "style": {
          marginTop: 24
        },
        "type": "primary",
        "onClick": this.onSetShowState
      }, "设置"), React.createElement(Button, {
        "style": {
          marginTop: 24
        },
        "type": "primary",
        "onClick": this.onSetShowTeamAB
      }, "设置TeamAB"))))));
    }
  });

  module.exports = App;

}).call(this);
