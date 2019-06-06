nullfunc = () ->
  return

AtcTitle = React.createClass
  componentDidMount: ->

  render: ()->
      <Navbar brand='Taiyouxi-admin' inverse toggleNavKey={0}>
        <Nav right eventKey={0}> {## This is the eventKey referenced ##}
          <NavItem eventKey={1} href='#'>Link</NavItem>
          <NavItem eventKey={2} href='#'>Link</NavItem>
          <DropdownButton eventKey={3} title='Dropdown'>
            <MenuItem eventKey='1'>Action</MenuItem>
            <MenuItem eventKey='2'>Another action</MenuItem>
            <MenuItem eventKey='3'>Something else here</MenuItem>
            <MenuItem divider />
            <MenuItem eventKey='4'>Separated link</MenuItem>
          </DropdownButton>
        </Nav>
      </Navbar>

AtcRestClient = (endpoint) ->
    _add_ending_slash = (string) ->
        string += "/"  unless string[string.length - 1] is "/"
        return string

    @endpoint = "/api/v1/"

    @api_call = (method, urn, callback, data) ->
      urn = _add_ending_slash(urn)
      console.log data
      $.ajax
        url: @endpoint + urn + endpoint
        dataType: "json"
        type: method
        data: data and JSON.stringify(data)
        contentType: "application/json; charset=utf-8"
        complete: (xhr, status) ->
            rc = {
                status: xhr.status
                json: xhr.responseJSON    
            }       
            callback rc  if callback?

    @shape = (callback, data) ->
      @api_call "POST", "shape", callback, data

    @unshape = (callback, data) ->
      @api_call "DELETE", "shape", callback

    @getCurrentShaping = (callback) ->
      @api_call "GET", "shape", callback

    @getToken = (callback) ->
      @api_call "GET", "token", callback

    @getAuthInfo = (callback) ->
      @api_call "GET", "auth", callback

    @updateAuthInfo = (address, data, callback) ->
      @api_call "POST", "auth/".concat(address), callback, data

    @mkSetting = (setting) ->
      return {      
          "down": {
            "loss": {
              "percentage": parseInt(setting.down.loss.percentage, 10),
              "correlation": parseInt(setting.down.loss.correlation, 10)
            },
            "delay": {
              "delay": parseInt(setting.down.delay.delay, 10),
              "jitter": parseInt(setting.down.delay.jitter, 10),
              "correlation": parseInt(setting.down.delay.correlation, 10)
            },
            "rate": parseInt(setting.down.rate, 10),
            "iptables_options": [],
            "corruption": {
              "percentage": parseInt(setting.down.corruption.percentage, 10),
              "correlation": parseInt(setting.down.corruption.correlation, 10)
            },
            "reorder": {
              "percentage": parseInt(setting.down.reorder.percentage, 10),
              "correlation": parseInt(setting.down.reorder.correlation, 10),
              "gap": parseInt(setting.down.reorder.gap, 10)
            }
        },
        "up": {
          "loss": {
            "percentage": parseInt(setting.up.loss.percentage, 10),
            "correlation": parseInt(setting.up.loss.correlation, 10)
          },
          "delay": {
            "delay": parseInt(setting.up.delay.delay, 10),
            "jitter": parseInt(setting.up.delay.jitter, 10),
            "correlation": parseInt(setting.up.delay.correlation, 10)
          },
          "rate": parseInt(setting.up.rate, 10),
          "iptables_options": [],
          "corruption": {
            "percentage": parseInt(setting.up.corruption.percentage, 10),
            "correlation": parseInt(setting.up.corruption.correlation, 10)
          },
          "reorder": {
            "percentage": parseInt(setting.up.reorder.percentage, 10),
            "correlation": parseInt(setting.up.reorder.correlation, 10),
            "gap": parseInt(setting.up.reorder.gap, 10)
          }
        }
      }

    return @


AtcSettings = ->
  @defaults =
      up:
          rate: null
          delay:
            delay: 0
            jitter: 0
            correlation: 0

          loss:
            percentage: 0
            correlation: 0

          reorder:
            percentage: 0
            correlation: 0
            gap: 0

          corruption:
            percentage: 0
            correlation: 0

          iptables_options: Array()

      down:
          rate: null
          delay:
            delay: 0
            jitter: 0
            correlation: 0

          loss:
            percentage: 0
            correlation: 0

          reorder:
            percentage: 0
            correlation: 0
            gap: 0

          corruption:
            percentage: 0
            correlation: 0

          iptables_options: Array()

  @getDefaultSettings = ->
    $.extend true, {}, @defaults

  @mergeWithDefaultSettings = (data) ->
    $.extend true, {}, @defaults, data

  return @


AtcProfileSelector = React.createClass
  getInitialState: () ->
        #TODO Curr Setting
        return {
            data : @props.data
            select_profile : null
            profiles:[]
        }

  selectProfile : (name) ->
    return () =>
        console.log "selectProfile"
        console.log name
        @setState {
            select_profile : name

        }

  getItemStyle : (data) ->
    if @state.select_profile? and @state.select_profile.name is data.name
        return true
    else
        return false

  addOneListItem : (data) ->
    n = data.name
    <ListGroupItem 
        onClick = {@selectProfile(data)} 
        active = {@getItemStyle(data)}
        key={n} >{n}</ListGroupItem>

  getAllProfileToSelect: () ->
    <div>
      <h6>Please Select A Profile :</h6>
      <ListGroup>
        { @state.profiles.map @addOneListItem }
      </ListGroup>
    </div>

  GetSelectedProfile : () -> @state.select_profile


  componentDidMount: () ->
    $.get "../api/v1/profiles/", (result) =>
      @setState {
            data           : @props.data
            select_profile : null
            profiles       : JSON.parse(result)
      }

  render:() ->
    @getAllProfileToSelect()
    



IPInput = React.createClass
  getInitialState : () ->
        {
          value: ''
        }
  GetIP   : () -> @state.value
  IsRight : () -> @validationState() is 'success'

  validationState : () ->
        exp=/^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/
        reg = @state.value.match(exp);
        if reg != null 
            return 'success'
        else
            return 'error'

  handleChange : () ->
    @setState {
      value: @refs.input.getValue()
    }
    @props.can_cb(@refs.input.getValue())

  render : () ->
      <Input
        type='text'
        value={@state.value}
        placeholder='Enter Client IP'
        label='Please Enter New Client IP'
        help='Validation is based on string in XXX.XXX.XXX.XXX.'
        bsStyle={@validationState()}
        hasFeedback
        ref='input'
        groupClassName='group-class'
        labelClassName='label-class'
        onChange={@handleChange} />

AtcClientModifyModal  = React.createClass
  getInitialState: () ->
        {
            data : @props.data
        }

  montify: () ->
    profile_data = @refs.selector1.GetSelectedProfile()
    console.log profile_data
    @props.setCb profile_data.content
    @props.onRequestHide()
    return

  render:() ->
      <Modal {...this.props} title={@props.name} animation={true}>
        <div className='modal-body'>
          <h4>Modify {@state.data.ip}</h4>
          <hr />
          <AtcProfileSelector ref="selector1"/>
        </div>
        <div className='modal-footer'>
            <ButtonToolbar>
                <Button
                    bsStyle='primary'
                    onClick={@montify}
                >
                    Modify
                </Button>
                <Button onClick={@props.onRequestHide}>Close</Button>
            </ButtonToolbar>
        </div>
      </Modal>


AtcClientNewModal  = React.createClass
  getInitialState : () ->
        {
          ip :null
          is_can_create : false
        }

  onCreateFinish : (nip) ->
    @props.update_cb(nip)
    @props.onRequestHide()
    return

  create: () ->
    setting = @refs.selector.GetSelectedProfile()
    client = new AtcRestClient(@state.ip)
    if setting isnt null
      setting = client.mkSetting setting.content
      client.shape @onCreateFinish.bind(@, @state.ip), setting
    else
      @onCreateFinish @state.ip
    return

  set_can_create: (nip) ->
        @setState {
          ip : nip
          is_can_create: true
        }

  render:() ->    
      <Modal {...this.props} title={@props.name} animation={true}>
        <div className='modal-body'>
          <IPInput can_cb = @set_can_create />
          <hr />
          <AtcProfileSelector ref="selector"/>
        </div>
        <div className='modal-footer'>
            <ButtonToolbar>
                <Button
                    bsStyle='primary'
                    onClick= @create
                    disabled={ !@state.is_can_create }
                >
                    Create
                </Button>
                <Button onClick={@props.onRequestHide}>Close</Button>
            </ButtonToolbar>
        </div>
      </Modal>

AtcClientContorl = React.createClass
  render:() ->
      <ButtonToolbar>
          <ModalTrigger modal={ <AtcClientNewModal data={ip:"10.0.1."} name = "Atc New" update_cb = @props.update_cb /> }>
            <Button bsStyle='primary'>New</Button>
          </ModalTrigger>
      </ButtonToolbar>


AtcClientModify = React.createClass
  getInitialState: () ->
        set = new AtcSettings()
        return {
            client : new AtcRestClient(@props.data.ip)
            is_turn_on : null,
            is_waiting : false
            settings   : set.defaults
        }

  initByCurrentShaping : (data) ->
    if data.json.down? and data.json.up?
      @setState {
        is_turn_on : "1"
        settings :  @state.client.mkSetting(data.json)
      }

  loadData : () ->
    @state.client.getCurrentShaping @initByCurrentShaping.bind(@)

  componentDidMount: () ->
    @state.client.getCurrentShaping @initByCurrentShaping.bind(@)

  turn: (rc) ->
      is_on = if @state.is_turn_on? then null else '1'
      @setState {
        is_turn_on : is_on
        is_waiting : false
      }
      return 

  changeSetting: (setting) ->
      console.log "ssss"
      console.log setting
      set = @state.client.mkSetting(setting)
      console.log "ssss2"
      console.log set
      @setState {
        settings : set
      }
      console.log "state"
      console.log @state
      console.log "mk"
      console.log {down: set.down, up: set.up}
      @state.client.shape nullfunc, {down: set.down, up: set.up}
      @setState {
        is_turn_on : '1'
        is_waiting : false
      }
      return 

  onClickTurn : () ->
      @setState { 
        is_turn_on : @state.is_turn_on
        is_waiting : true 
      }
      if !@state.is_turn_on
          @state.client.shape @turn.bind(@), {down: @state.settings.down, up: @state.settings.up}
      else
          @state.client.unshape @turn.bind(@)
      return

  getTurnText : (is_waiting, is_turn_on) ->
        if is_waiting
            return "waiting ..."
        else 
            if is_turn_on then "Turn Off" else "Turn On"

  render:() ->
      is_waiting = @state.is_waiting
      is_turn_on = @state.is_turn_on
      <ButtonToolbar>
        <ModalTrigger modal={
          <AtcClientModifyModal 
            data        = { @props.data } 
            setCb       = @changeSetting
            currSetting = { @state.settings }
            name = "Atc Modify"/>}>
            <Button bsStyle='danger'>Modify</Button>
        </ModalTrigger>
        <Button 
            bsStyle  = { if is_turn_on then 'success' else 'warning' }
            disabled = { is_waiting }
            onClick  = @onClickTurn
        >
            { @getTurnText(is_waiting,is_turn_on) }
        </Button>
      </ButtonToolbar>


AtcClientList = React.createClass
  getInitialState: () ->
    {
        atc_data_list: []
    }

  loadData : (nip) ->
    $.get "../api/v1/clientall/", (result) =>
      list = {}

      for a in @state.atc_data_list
        list[a.ip] = a

      if nip?
          list[nip] = {
            ip:nip
          }
      res = JSON.parse(result)

      for r in res
        list[r.ip] = r

      new_list = []

      for k,v of list
        new_list.push v

      @setState {
            atc_data_list : new_list
      }

  componentDidMount: () -> @loadData()

  AddOneToTable : ( data ) ->
    <tr key={data.ip} >
        <td>{data.name}</td>
        <td>{data.ip}</td>
        <td><AtcClientModify data = {data}/></td>
    </tr>

  AddClient : (nip) ->
    atc_data_list = @state.atc_data_list
    atc_data_list[atc_data_list.length] = {
      ip : nip,
      idx : atc_data_list.length
    }
    @setState {
            atc_data_list       : atc_data_list
    }

  getTables : () ->
    results = @state.atc_data_list
    <tbody>
        { results.map @AddOneToTable }
    </tbody>

  render:() ->
      <Table striped bordered condensed hover>
        <thead>
          <tr>
            <th>Name</th>
            <th>IP</th>
            <th>Op</th>
          </tr>
        </thead>
        { @getTables() }
      </Table>


AtcAdmin = React.createClass
  getInitialState: () -> {}

  update_all : (nip) ->
    @refs.list.loadData(nip)

  render: ()->
    <div>
      <AtcTitle         ref = "title"   update_cb = @update_all.bind(@) />
      <AtcClientContorl ref = "contorl" update_cb = @update_all.bind(@) />
      <AtcClientList    ref = "list"    update_cb = @update_all.bind(@) />
    </div>

React.render <AtcAdmin />, document.getElementById('all')