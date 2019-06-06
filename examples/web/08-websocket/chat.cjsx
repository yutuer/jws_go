ChatClient = React.createClass
  getInitialState : () ->
        {
          room_key :"0"
          state : "init"
          chat_history: []
          value : "Enter..."
        }

  componentDidMount: () -> 
    if window["WebSocket"]? 
        conn = new WebSocket("ws://127.0.0.1:7788/ws")
        conn.onclose = (evt) =>
            @setState {
              state : "connect_close"
            }
        conn.onmessage = (evt) => @addMsg evt.data

        @setState {
          conn : conn
        }
    else
      @setState {
        state : "NoWebSocket"
      }

    return

  addMsg : (msg) ->
    his = @state.chat_history
    his.push {
      sender : "fy"
      text   : msg
    }
    @setState {
      chat_history : his
    }

  mkHistoryItem: (msg) ->
    <div id="msg">
      <p> <strong> {msg.sender} </strong>  {msg.text} </p>
    </div>

  getHistory : () ->
    return @state.chat_history.map @mkHistoryItem

  sendMsg : () ->
    console.log @refs.msg_input.getValue()
    msg = @refs.msg_input.getValue()
    @state.conn.send(msg);

  render:() ->
      <div>
        <p id="history" ref="history" >{@getHistory()}</p>
        <div>
            <div id="sender">
                <Input type='text' ref='msg_input' />
                <Button bsStyle='primary' onClick={@sendMsg}>
                    Send
                </Button>
                <Button bsStyle='primary' onClick={@sendMsg}>
                    Clean
                </Button>
            </div>
        </div>
      </div>

React.render <ChatClient />, document.getElementById('chater')