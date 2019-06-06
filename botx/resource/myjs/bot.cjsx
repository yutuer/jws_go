
submit_get = (dataString) ->
    $.ajax({
      type: "GET",
      url: "addbot",
      data: dataString,
      dataType: "json",
      success:  (data) ->
           $("#forminfo").html("<div class=\"alert alert-success\" role=\"alert\">" +"500 bot added. Everyting goes ("+data.status+")" +"</div>").show().fadeOut(10000);
    });


SimpleGenerator = React.createClass
    getInitialState: () -> {
        botnumber:500
        m:"SimpleGenerator"
    }

    handleChange: (field, e) ->
        nextState={}
        nextState[field] = e.target.value
        @setState(nextState)

    handleSubmit:() ->
        submit_get @state

    render: () ->
        <div>
            <label htmlFor="sg-number" className="col-sm-4 control-label">机器人总数</label>
            <div className="col-sm-8">
                <input id="sg-number" className="form-control" ref="botnumber" defaultValue={@state.botnumber} name="number" type="number"  min="1" onChange={@handleChange.bind(@, 'botnumber')}></input>
            </div>
            <button type="submit" className="btn btn-default" onClick={@handleSubmit}>Submit</button>
        </div>

RandomGenerator = React.createClass
    getInitialState: () -> {
        botnumber:500
        oncenumber:10
        sleep:5
        m: "RandomGenerator"
    }

    botnumberChange:(event)->
            @setState {
                botnumber: event.target.value
            }

    oncenumberChange:(event)->
            @setState {
                oncenumber: event.target.value
            }

    sleepChange:(event)->
            @setState {
                sleep: event.target.value
            }

    handleSubmit:() ->
        submit_get @state

    render: () ->
        <div>
            <div className="form-group">
                <label htmlFor="rg-number" className="col-sm-4 control-label">机器人总数</label>
                <div className="col-sm-8">
                    <input id="rg-number" className="form-control" ref="botnumber" defaultValue={@state.botnumber} name="number" type="number" min="1" onChange={@botnumberChange}></input>
                </div>
            </div>
            <div className="form-group">
                <label htmlFor="rg-oncenumber" className="col-sm-4 control-label">机器人单次最大产量</label>
                <div className="col-sm-8">
                    <input id="rg-oncenumber" className="form-control" ref="oncenumber" defaultValue={@state.oncenumber} name="oncenumber" type="number" min="1" onChange={@oncenumberChange}></input>
                </div>
            </div>
            <div className="form-group">
                <label htmlFor="rg-sleep" className="col-sm-4 control-label">Sleep范围</label>
                <div className="col-sm-8">
                    <input id="rg-sleep" className="form-control" ref="sleep" defaultValue={@state.sleep} name="sleep" type="number" min="2" onChange={@sleepChange}></input>
                </div>
            </div>
            <button type="submit" className="btn btn-default" onClick={@handleSubmit}>Submit</button>
        </div>


BotGenerator = React.createClass
    getInitialState: () -> {
        n: {
            SimpleGenerator:<SimpleGenerator/>
            RandomGenerator:<RandomGenerator/>
        }
    }
    make: () ->
        $.map @state.n, (v, key)->
            <TabPane key={key} eventKey={key} tab={key}>
              {v}
            </TabPane>
    render: () ->
        <TabbedArea defaultActiveKey="SimpleGenerator">
            {@make()}
        </TabbedArea>

React.render <BotGenerator />, document.getElementById('reactbot')
