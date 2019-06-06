React = require 'react'
$     = require 'jquery'

AtcProfileList = React.createClass
  getInitialState: () ->
    {
      profiles:[]
    }

  loadAllInfo: () ->

  addOneToTable : ( data ) ->
    <tr key={data.id}>
        <td>{data.id}</td>
        <td>{data.name}</td>
        <td>{data.content.up.rate}</td>
        <td>{data.content.up.delay.delay}</td>
        <td>{data.content.up.loss.percentage}</td>
        <td>{data.content.up.corruption.percentage}</td>
        <td>{data.content.up.reorder.percentage}</td>
        <td>{data.content.down.rate}</td>
        <td>{data.content.down.delay.delay}</td>
        <td>{data.content.down.loss.percentage}</td>
        <td>{data.content.down.corruption.percentage}</td>
        <td>{data.content.down.reorder.percentage}</td>
    </tr>

  getTables : () ->
    results = @state.profiles
    <tbody>
        { results.map @addOneToTable }
    </tbody>

  componentDidMount: () ->
    $.get "../api/v1/profiles/", (result) =>
      @setState {
            profiles       : JSON.parse(result)
      }

  render:() ->
      @loadAllInfo()
      <Table striped bordered condensed hover>
        <thead>
          <tr>
            <th>Index</th>
            <th>Name</th>
            <th>Up Bandwidth</th>
            <th>Up Latency</th>
            <th>Up Loss</th>
            <th>Up Corruption</th>
            <th>Up Reorder</th>
            <th>Down Bandwidth</th>
            <th>Down Latency</th>
            <th>Down Loss</th>
            <th>Down Corruption</th>
            <th>Down Reorder</th>
          </tr>
        </thead>
        { @getTables() }
      </Table>
