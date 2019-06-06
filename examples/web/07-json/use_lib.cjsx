reqJson = (num, cb) ->
      $.ajax {
        type: "GET",
        url: "../api/v1/get",
        data: "num="+num,
        dataType: "text",
        success:  cb
      }

JsonModByUrl = React.createClass
  getInitialState: () ->
    {
      editor:null
    }

  componentDidMount: () ->
    container = document.getElementById(@props.name + "jsoneditor");
    change_cb = @onchange.bind @
    options = {
      "mode": "tree",
      "search": true,
      change : change_cb
    }
    editor_imp = new JSONEditor(container, options)
    @getData (data) -> 
        editor_imp.set(JSON.parse (data))
        change_cb(data)
        return
    @setState {
      editor:editor_imp
    }
    return

  getData: (cb) ->
    reqJson 22, cb
    return

  onchange : () ->
    data = @state.editor.getText()
    @props.onchange(data)

  render:() ->
        <div id={@props.name + "jsoneditor"} />

# New Json Editor With Shown

JsonEditor = React.createClass
  getInitialState: () ->
    {
      json : "{}"
    }

  componentDidMount: () ->
    return

  onchange:(data) ->
    console.log "onchange"
    console.log data
    @setState {
        json : data
    }
    console.log @state

  render:() ->
    <div> 
      <input type="text" value={@state.json}/>
      <JsonModByUrl id={name} key={name} name={name} onchange={@onchange}/> 
    </div>



React.render <JsonEditor />, document.getElementById('example4')