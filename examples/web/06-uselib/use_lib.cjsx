JsonMod = React.createClass
  getInitialState: () ->
    {
    }

  componentDidMount: () ->
    container = document.getElementById(@props.name + "jsoneditor");
    options = {
      "mode": "tree",
      "search": true
    }
    editor = new JSONEditor(container, options)
    editor.set(JSON.parse (@props.data))
    return

  render:() ->
        <div id={@props.name + "jsoneditor"} />


mkJsonMod = (data, name) ->
  <div> 
    <JsonMod id={name} key={name} data={data} name={name} /> 
  </div>

json = {
    "Array": [1, 2, 3],
    "Boolean": true,
    "Null": null,
    "Number": 123,
    "Object": {"a": "b", "c": "d"},
    "String": "Hello World"
};

js = mkJsonMod JSON.stringify(json), "test"

React.render js, document.getElementById('example3')