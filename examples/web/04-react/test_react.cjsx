Test01 = React.createClass
  getInitialState : () ->
        {
          name :"0"
        }

  componentDidMount: () -> 
        @setState {
          name :@props.name
        }

  changeName: () ->
    new_name = "R" + @state.name
    console.log new_name
    @setState {
      name: new_name
    }

  render:() ->
      <div>
        <hr />
        <h1>Hello, {@state.name}!</h1>
        <input id="t2" type="text" name="lastname" />
        <input type="button" value="Focus" onClick={@changeName} />
      </div>

React.render <Test01 name="aaaaaa" />, document.getElementById('example2')