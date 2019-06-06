DefaultRoomKey = "all"

ws_host = "ws://10.0.1.221:10010/ws"

ChatClient = React.createClass
  getInitialState : () ->
        u = window.location.href
        pre = "http://"
        e = u.indexOf("/", pre.length)
        ws_host = "ws://" + u.substring(pre.length, e) + "/ws"
        {
          room_key :"0"
          state : "init"
          chat_history: []
          value : "Enter..."
        }

  componentDidMount: () -> 
    if window["WebSocket"]? 
        conn = new WebSocket(ws_host)
        conn.onclose = (evt) =>
            @setState {
              state : "connect_close"
            }
        conn.onmessage = (evt) => @addMsg false, evt.data
        conn.onopen = () =>
          conn.send("Reg:" + DefaultRoomKey + ":WebChater"+@guid())

        @setState {
          conn : conn
        }
    else
      @setState {
        state : "NoWebSocket"
      }

    return

  addMsg : (is_myself, msg) ->
    his = @state.chat_history
    his.push {
      sender : if is_myself then "Me:" else "Other:"
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
    msg = @refs.msg_input.getValue()
    @state.conn.send("Say:" + DefaultRoomKey + ":" + msg)
    @addMsg true, msg

  guid : () ->
    s4 = =>
      Math.floor((1 + Math.random()) * 0x10000).toString(16).substring(1)
    return s4() + s4() + '-' + s4() + '-' + s4() + '-' + s4() + '-' + s4() + s4() + s4();


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