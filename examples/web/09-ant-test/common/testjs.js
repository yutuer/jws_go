import {Datepicker, message} from 'antd';

var App = React.createClass({
  getInitialState() {
    return {
      date: ''
    };
  },
  handleChange(value) {
    this.setState({
      date: value
    });
  },
  notice() {
    message.info(this.state.date.toString());
  },
  render() {
    return <div>
      <Datepicker onSelect={this.handleChange} />
      <button className="ant-btn ant-btn-primary" onClick={this.notice}>显示日期</button>
    </div>;
  }
});
module.exports = App;