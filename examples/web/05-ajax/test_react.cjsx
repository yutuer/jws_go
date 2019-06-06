reqJson = (num, cb) ->
      $.ajax {
        type: "GET",
        url: "../api/v1/get",
        data: "num="+num,
        dataType: "text",
        success:  cb
      }



Test01 = React.createClass
  getInitialState : () ->
        {
          name :"0"
          data: []
        }

  componentDidMount: () -> 
      reqJson 22, (result) =>
              re = JSON.parse(result)
              @setState {
                    name : @props.name
                    data : re.data
              }

  changeName: () ->
    ndata = @state.data
    new_item = "TestItem" + (ndata.length + 1).toString()
    ndata.push new_item
    @setState {
      name: new_item
      data: ndata
    }

  addOneToTable : ( data ) ->
      <tr key={data}>
          <td>{data}</td>
          <td>{data}</td>
      </tr>

  getTables : () ->
    <tbody>
      {@state.data.map @addOneToTable}
    </tbody>

  render:() ->
      <div>
        <hr />
        <h1>Hello, {@state.name}!</h1>
        <input id="t2" type="text" name="lastname" />
        <input type="button" value="Focus" onClick={@changeName} />

        <a href="http://127.0.0.1:7788/api/v1/get?num=99">Ajax请求</a>
        <table>
            <tr>
              <th>Name</th>
              <th>V</th>
            </tr>
            { @getTables() }
        </table>
      </div>

React.render <Test01 name="aaaaaa" />, document.getElementById('example2')

