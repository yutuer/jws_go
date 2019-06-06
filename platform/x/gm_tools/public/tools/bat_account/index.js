
(function () {
    var App, React, antd, Icon, Upload, Button, message, BatchAccountResult;

    React = require('react');

    antd = require('antd');

    Icon = antd.Icon;

    Upload = antd.Upload;

    Button = antd.Button;

    message = antd.message;

    BatchAccountResult = require('./bat_account_result');

    App = React.createClass({displayName: "App",
        getInitialState: function () {
            return {
                batch_query_index: "",
                batch_index: ""
            };
        },

        openExcel:function (data) {
            console.log(data)
            if (data.file.status === "done") {
                message.success(data.file.name + ' 上传成功。');
                this.refs.batch_result.Refresh(data.file.response);

            }
            if (data.file.status === "error") {
                return alert(data.file.response);
            }
        },


        onChange: function(e) {
            return this.setState({
                batch_query_index: e.target.value
            });
        },
        render:function () {
            return React.createElement("div", null, React.createElement(Upload, Object.assign({}, this.props, {
                "action": "/api/v1/banAccountsAndDelRanks",
                "onChange": this.openExcel
            }), React.createElement(Button, {
                "type": "ghost"
            }, React.createElement(Icon, {
                "type": "upload"
            }), " 批量封禁")),React.createElement("hr", null), React.createElement("hr", null), React.createElement("div", null, React.createElement(BatchAccountResult, {
                    "ref": "batch_result"
                }))
            )
            
        }
    });

    module.exports = App;


}).call(this);