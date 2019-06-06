React = require 'react'
$     = require 'jquery'
GameServerList = React.createClass
  getInitialState: () ->
        #TODO Curr Setting
        return {
            data : @props.data
            select_profile : null
            servers:[]
        }

  selectProfile : (name) ->
    return () =>
        @setState {
            select_profile : name

        }

  getItemStyle : (data) ->
    if @state.select_profile is data
        return true
    else
        return false

  addOneListItem : (data) ->
    n = data
    <ListGroupItem
        onClick = {@selectProfile(data)}
        active = {@getItemStyle(data)}
        key={n} >{n}</ListGroupItem>

  getAllProfileToSelect: () ->
    <div>
      <h6>Please Select A Profile :</h6>
      <ListGroup>
        { @state.servers.map @addOneListItem }
      </ListGroup>
    </div>

  GetSelectedProfile : () -> @state.select_profile


  componentDidMount: () ->
    $.get "../api/v1/server_all", (result) =>
      @setState {
            servers : JSON.parse(result)
      }

  render:() ->
    @getAllProfileToSelect()


AccountKeyInput = React.createClass
  getInitialState : () ->
        {
          value: ''
        }

  IsRight : () -> @validationState() is 'success'

  validationState : () ->
        exp=/^(\w+)\:(\d+)\:(\d+)\:(\d+)$/
        exp_init=/^init:(\w+)\:(\d+)\:(\d+)\:(\d+)$/
        if not @refs.input?
          return 'error'
        v =  @refs.input.getValue()
        reg = v.match(exp);
        reg_init = v.match(exp_init);

        if reg != null or reg_init != null
            return 'success'
        else
            return 'error'

  handleChange : () ->
    @setState {
      value: @refs.input.getValue()
    }
    @props.can_cb(@refs.input.getValue(), @IsRight())

  render : () ->
      <Input
        type='text'
        value={@state.value}
        placeholder='Enter Account Key'
        label='Please Enter Account Key'
        help='Validation is based on string in profile:X:X:XXX'
        bsStyle={@validationState()}
        hasFeedback
        ref='input'
        groupClassName='group-class'
        labelClassName='label-class'
        onChange={@handleChange} />

# props name old [min max) can_cb()
NumInput = React.createClass
  getInitialState : () ->
        {
          value: parseInt(@props.old)
        }

  IsRight : () -> @validationState() is 'success'

  validationState : () ->
        exp=/^(\d+)$/
        if not @refs.input?
          return 'error'
        v =  @refs.input.getValue()
        reg = v.match(exp);

        if reg != null &&
           parseInt(@props.min) <= parseInt(v) < parseInt(@props.max)
            return 'success'
        else
            return 'error'

  handleChange : () ->
    @setState {
      value: @refs.input.getValue()
    }
    v = @refs.input.getValue()
    @props.can_cb(parseInt(v), @IsRight())

  render : () ->
      <Input
        type='text'
        value={@state.value}
        placeholder={'Enter ' + @props.name}
        label={'Please Enter ' + @props.name}
        help={'MUST in [' + @props.min + "," + @props.max + ")"}
        bsStyle={@validationState()}
        hasFeedback
        ref='input'
        groupClassName='group-class'
        labelClassName='label-class'
        onChange={@handleChange} />


getAvataName = (id) ->
  names = ["关羽", "张飞", "孙尚香"]
  n = names[id]
  if not n?
    n = "未定义主角" + id.toString()
  return n

getScName = (id) ->
  names = ["金钱", "精铁"]
  n = names[id]
  if not n?
    n = "未使用类型" + id.toString()
  return n

getHcName = (id) ->
  names = ["购买钻", "赠送钻", "补偿钻", "钻石总量"]
  n = names[id]
  if not n?
    n = "未使用类型" + id.toString()
  return n

mkModifyModel = (logic) -> React.createClass
  getInitialState: () ->
        {
            cb : @props.cb
            maker : @props.maker
            is_ok : true
            v : @props.v
            l : logic
        }

  montify: () ->
    console.log @state
    @props.cb @props.maker.key, @state.v
    @props.onRequestHide()
    return

  render:() ->
      <Modal {...this.props} title={@props.maker.name} animation={true}>
        <div className='modal-body'>
          { @state.l.ModifyRender(@) }
          <ButtonToolbar>
                <Button
                    bsStyle='primary'
                    onClick={@montify}
                    disabled={!@state.is_ok}
                >
                    Montify
                </Button>
                <Button onClick={@props.onRequestHide}>Cancel</Button>
          </ButtonToolbar>
        </div>
      </Modal>


JsonMod = React.createClass
  getInitialState: () ->
    {
    }

  componentDidMount: () ->
    container = document.getElementById(@props.name + "jsoneditor");
    options = {
        mode: 'view'
    }
    editor = new JSONEditor(container, options)
    editor.set(JSON.parse (@props.data))
    return

  render:() ->
        <div id={@props.name + "jsoneditor"} >
        </div>


mkJsonMod = (data, name) ->
  <div>
    <JsonMod id={name} key={name} data={data} name={name} />
  </div>




SendToAllMailReward = React.createClass
  getInitialState: () ->
        {
        }

  montify: () ->
    console.log "montify"
    @props.onRequestHide()
    return

  render:() ->
      <Modal {...this.props} title="全服发送奖励邮件" animation={true}>
        <div className='modal-body'>
          <ButtonToolbar>
                <hr />
                <Button
                    bsStyle='primary'
                    onClick={@montify}
                >发送</Button>
                <Button onClick={@props.onRequestHide}>取消</Button>
          </ButtonToolbar>
        </div>
      </Modal>




MailAllSender = React.createClass
  getInitialState: () ->
    {
      servers:[]
    }

  componentDidMount: () ->
    $.get "../api/v1/server_all", (result) =>
      @setState {
            servers : JSON.parse(result)
      }

  render:() ->
        <div>
            <div>
                <div id="header">
                    <h1>"MailAllSender"</h1>
                </div>
            </div>
            <div>
                <div id="servers">
                    <GameServerList ref="selector" />

                </div>
                <ModalTrigger modal={<SendToAllMailReward />}>
                  <Button bsStyle='primary' onClick=null >Modify</Button>
                </ModalTrigger>
            </div>
        </div>
