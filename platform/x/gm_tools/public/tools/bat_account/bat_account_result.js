
(function () {
    var $, React, boot, Table, Label, Loading, BatchAccountResult;

    React = require('react');

    Loading = require('react-loading');

    boot = require('react-bootstrap');

    Table = boot.Table;

    Label = boot.Label;

    $ = require('jquery');



    BatchAccountResult = React.createClass({displayName: "BatchAccountResult",
        ResultMap: {
            "1": "时间格式错误",
            "2": "奖励格式错误",
            "3": "ID必须是整数",
            "4": "不存在的服务器",
            "5": "服务器发送失败",
            "6": "uid错误"
        },
        getInitialState: function() {
            return {
                count: 0,
                mails: []
            };
        },
        Refresh: function(data) {
            console.log(data);
            return this.setState({
                mails: data.FailArr,
                count: data.SuccessCount
            });
        },
        getAllInfo: function(data) {
            var k, re, v;
            if (data == null) {
                return React.createElement("div", null, "UnKnown Info");
            }
            re = [];
                console.log(re);
            for (k in data) {
                v = data[k];
                console.log(k);
                re.push(React.createElement("tr", {
                    "key": k
                }, React.createElement("td", null, v.Id), React.createElement("td", null, React.createElement(Label, {
                    "bsStyle": "danger"
                }, "失败")), React.createElement("td", null, this.ResultMap[v.FailCode])));
            }
            return re;
        },
        getResultList: function() {
            if (this.state.is_loading) {
                return React.createElement(Loading, {
                    "type": 'spin',
                    "height": "200",
                    "width": "200",
                    "color": '#61dafb'
                });
            }
            return React.createElement(Table, {
                "striped": true,
                "bordered": true,
                "condensed": true,
                "hover": true
            }, React.createElement("thead", null, React.createElement("tr", null, React.createElement("th", null, "ID"), React.createElement("th", null, "结果"), React.createElement("th", null, "信息"))), React.createElement("tbody", null, this.getAllInfo(this.state.mails)));
        },
        render: function() {
            return React.createElement("div", {
                "className": "ant-form-inline"
            }, React.createElement("div", null, React.createElement(Label, {
                "bsStyle": "success"
            }, "成功：", this.state.count), React.createElement(Label, {
                "bsStyle": "danger"
            }, "失败：", (Object.keys(this.state.mails).length))), React.createElement("hr", null), this.getResultList());
        }
    });

    module.exports = BatchAccountResult;


}).call(this);