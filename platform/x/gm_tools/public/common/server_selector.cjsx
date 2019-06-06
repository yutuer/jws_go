antd   = require 'antd'
React  = require 'react'
$      = require 'jquery'
Select = antd.Select
Option = Select.Option

ServerSelector = React.createClass
  getInitialState: () ->
        return {
            servers : []
            servers_cfg : {}
        }

  handleChange : (name) ->
    @props.handleChange(@state.servers_cfg[name])

  addOneServerOption : (data) ->
        name = data.name.toString()
        return <Option key={name} value={name}>{name}</Option>

  getAllProfileToSelect: () ->
        <Select  showSearch={false}  style={{width:300}} onChange={@handleChange}>
          { @state.servers.map @addOneServerOption }
        </Select>

  componentDidMount: () ->
    $.get "../api/v1/server_cfg_all", (result) =>
        serv = JSON.parse(result)
        cfg = {}
        for s in serv
          cfg[s.name] = s

        @setState {
            servers     : serv
            servers_cfg : cfg
        }

  render:() ->
    @getAllProfileToSelect()


ServerPicker = React.createClass
  getInitialState: () ->
        return {
            servers : []
        }

  handleChange : (data) ->
    @props.handleChange(data)

  addOneServerOption : (data) ->
        <Option value={data.name.toString()}>{data}</Option>

  getAllProfileToSelect: () ->
        <Select multiple showSearch={true}  style={{width:300}} onChange={@handleChange}>
          { @state.servers.map @addOneServerOption }
        </Select>

  componentDidMount: () ->
    $.get "../api/v1/server_cfg_all", (result) =>
        @setState {
            servers : JSON.parse(result)
        }

  render:() ->
    @getAllProfileToSelect()


module.exports = {
    ServerSelector : ServerSelector
    ServerPicker   : ServerPicker
}
