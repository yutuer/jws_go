// Generated by CoffeeScript 1.12.7
(function() {
  var Api, NewSendModal, NoticeTable, React, SysRollNotice, Table, TimeUtil, antd, boot;

  antd = require('antd');

  boot = require('react-bootstrap');

  Api = require('../api/api_ajax');

  TimeUtil = require('../util/time');

  SysRollNotice = require('./notice');

  Table = boot.Table;

  NewSendModal = require('./new_sys_roll_notice');

  React = require('react');

  NoticeTable = React.createClass({displayName: "NoticeTable",
    getInitialState: function() {
      return {
        notices: [],
        is_loading: true
      };
    },
    Refersh: function(time_wait, callback) {
      var all_notice, api, count, i, len, ref, results, server_id;
      this.setState({
        is_loading: true
      });
      if ((time_wait != null) && time_wait > 0) {
        setTimeout(this.Refersh.bind(this, null, callback), 1);
        return;
      }
      api = new Api();
      all_notice = [];
      count = 0;
      ref = this.props.server_id;
      results = [];
      for (i = 0, len = ref.length; i < len; i++) {
        server_id = ref[i];
        results.push(api.Typ("getSysRollNotice").ServerID(server_id).AccountID(this.props.account_id).Key(this.props.curr_key).Do((function(_this) {
          return function(result) {
            var k, ref1, v;
            count++;
            console.log(result);
            ref1 = JSON.parse(result);
            for (k in ref1) {
              v = ref1[k];
              all_notice[v.id] = new SysRollNotice(v);
            }
            if (count === _this.props.server_id.length) {
              _this.setState({
                notices: all_notice,
                is_loading: false
              });
              if (_this.props.on_loaded != null) {
                _this.props.on_loaded();
              }
              if (callback != null) {
                return callback();
              }
            }
          };
        })(this)));
      }
      return results;
    },
    getDelButton: function(is_get, id, server_id) {
      return React.createElement("delButton", {
        "id": id,
        "key": id,
        "on_deled": ((function(_this) {
          return function() {
            return _this.Refersh(1);
          };
        })(this)),
        "disabled": is_get,
        "server_id": server_id,
        "account_id": this.props.account_id
      });
    },
    del: function(id, server_id) {
      return (function(_this) {
        return function() {
          var api;
          api = new Api();
          return api.Typ("delSysRollNotice").ServerID(server_id).AccountID(_this.props.account_id).Key(_this.props.curr_key).Params(id).Do(function(result) {
            console.log("on_deled");
            return _this.Refersh(0.01);
          });
        };
      })(this);
    },
    activityPublic: function(id, server_id) {
      return (function(_this) {
        return function() {
          var s, states;
          states = _this.state.notices;
          s = states[id];
          s.ChangeState();
          states[id] = s;
          _this.setState({
            notices: states
          });
          return s.UpdateToServer(server_id, _this.props.account_id, _this.props.curr_key, function() {
            return _this.Refersh(0.01);
          });
        };
      })(this);
    },
    mod: function(id) {
      return (function(_this) {
        return function() {
          return _this.refs["new_mod_" + id].showModal();
        };
      })(this);
    },
    handleModOk: function(notice) {
      var refresh;
      console.log("handleModOk");
      console.log(notice);
      refresh = (function(_this) {
        return function() {
          return _this.Refersh(0.01);
        };
      })(this);
      return notice.UpdateToServer(notice.server_id, this.props.account_id, this.props.curr_key, refresh);
    },
    getStateString: function(v) {
      var now_t, v_begin_t, v_end_t;
      if (v.state === 0) {
        return "未发布";
      }
      now_t = new Date();
      v_begin_t = new Date(v.json.command.params[0]);
      v_end_t = new Date(v.json.command.params[1]);
      if ((v_begin_t <= now_t && now_t < v_end_t)) {
        return "已发布";
      }
      if (now_t >= v_end_t) {
        return "已过期";
      }
      if (now_t < v_begin_t) {
        return "未到期";
      }
    },
    getAllInfo: function(data) {
      var k, re, v;
      if (data == null) {
        return React.createElement("div", null, "UnKnown Info");
      }
      re = [];
      for (k in data) {
        v = data[k];
        re.push(React.createElement("tr", {
          "key": v.id
        }, React.createElement("td", null, v.id), React.createElement("td", null, v.server_id), React.createElement("td", null, v.interval), React.createElement("td", null, v.title), React.createElement("td", null, v.json.command.params[0]), React.createElement("td", null, v.json.command.params[1]), React.createElement("td", null, v.State()), React.createElement("td", {
          "className": 'row-flex row-flex-middle row-flex-start'
        }, React.createElement(NewSendModal.NewModal, Object.assign({}, this.props, {
          "ref": "new_mod_" + v.id,
          "notice": v,
          "server_id": [v.server_id],
          "on_ok": this.handleModOk
        })), React.createElement(boot.Button, {
          "bsStyle": 'success',
          "disabled": false,
          "onClick": this.activityPublic(v.id, v.server_id)
        }, "发送"), React.createElement(boot.Button, {
          "bsStyle": 'info',
          "disabled": false,
          "onClick": this.mod(v.id)
        }, "修改"), React.createElement(boot.Button, {
          "bsStyle": 'danger',
          "disabled": false,
          "onClick": this.del(v.id, v.server_id)
        }, "删除"))));
      }
      return re;
    },
    getRewardList: function() {
      if (this.state.is_loading) {
        return React.createElement("i", {
          "className": "anticon anticon-loading"
        });
      }
      return React.createElement(boot.Table, Object.assign({}, this.props, {
        "striped": true,
        "bordered": true,
        "condensed": true,
        "hover": true
      }), React.createElement("thead", null, React.createElement("tr", null, React.createElement("th", null, "公告ID"), React.createElement("th", null, "公告大区"), React.createElement("th", null, "循环间隔(秒)"), React.createElement("th", null, "公告注释"), React.createElement("th", null, "开始时间"), React.createElement("th", null, "结束时间"), React.createElement("th", null, "状态"), React.createElement("th", null, "操作"))), React.createElement("tbody", null, this.getAllInfo(this.state.notices)));
    },
    render: function() {
      return React.createElement("div", Object.assign({}, this.props, {
        "className": "ant-form-inline"
      }), this.getRewardList());
    }
  });

  module.exports = NoticeTable;

}).call(this);
