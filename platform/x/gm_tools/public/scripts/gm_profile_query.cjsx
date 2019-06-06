React = require 'react'
ReactBootstrap = require 'react-bootstrap'
$              = require 'jquery'

ButtonToolbar  = ReactBootstrap.ButtonToolbar
Button         = ReactBootstrap.Button
MenuItem       = ReactBootstrap.MenuItem
DropdownButton = ReactBootstrap.DropdownButton
Table          = ReactBootstrap.Table
ModalTrigger   = ReactBootstrap.ModalTrigger
Modal          = ReactBootstrap.Modal
Navbar         = ReactBootstrap.Navbar
Nav            = ReactBootstrap.Nav
NavItem        = ReactBootstrap.NavItem
Input          = ReactBootstrap.Input
ListGroup      = ReactBootstrap.ListGroup
ListGroupItem  = ReactBootstrap.ListGroupItem

console.group "profilemod"
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
        if @props.set_ser?
          @props.set_ser(name)

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
        exp=/^(\w+)\:(\d+)\:(\d+)\:(.+)$/
        exp_init=/^init:(\w+)\:(\d+)\:(\d+)\:(.+)$/
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
        help='Validation is based on string in X:X:X:XXX'
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


###################################################################
###
{
  "Avatars": [
    {
      "Level": 1,
      "Xp": 0
    },
    {
      "Level": 1,
      "Xp": 0
    },
    {
      "Level": 1,
      "Xp": 0
    }
  ]
}###
GameAvatarExpInfo = React.createClass
  getInitialState: () ->
    o = JSON.parse @props.data
    return {
      exps: o.Avatars
    }

  SetData : (data) ->
    o = JSON.parse data
    console.log o
    @setState {
      exps: o.Avatars
    }

  getAllInfo : (exps) ->
    if not exps?
      return <div>UnKnown Info</div>

    re = []
    for k, v of exps
      console.log v
      re.push <tr key={k} >
              <td>{k}</td>
              <td>{getAvataName(k)}</td>
              <td>{v.Level}</td>
              <td>{v.Xp}</td>
          </tr>
    return re

  render:() ->
      <Table striped bordered condensed hover>
        <thead>
          <tr>
            <th>AvatarId</th>
            <th>Name</th>
            <th>Level</th>
            <th>Xp</th>
          </tr>
        </thead>
        <tbody>
          { @getAllInfo(@state.exps) }
        </tbody>
      </Table>

GameAvatarExpModifyModal = mkModifyModel {
    Get : (v) ->
      obj = JSON.parse v
      return obj.Avatars

    GetOne : (v, typ, idx) ->
      obj = JSON.parse v
      return obj.Avatars[idx][typ]

    OnChange : (react_obj, typ, idx, v, ok) ->
      nv = JSON.parse react_obj.state.v
      nv.Avatars[parseInt(idx)][typ] = v
      react_obj.setState {
        is_ok : ok
        v : JSON.stringify nv
      }

    mkAvatarMods : (react_obj) ->
        self = @
        re = []
        avatar_exps = self.Get react_obj.state.v
        for k, avatar_exp of avatar_exps
          do (k, avatar_exp) ->
            re.push <div>
                <div>Mod {getAvataName(k)} Exp: </div>
                <hr />
                <NumInput
                    can_cb = { (v, ok) -> self.OnChange(react_obj, "Level", k, v, ok) }
                    name="Level"
                    old= { self.GetOne(react_obj.state.v, "Level", k ) }
                    min="1"
                    max="100" />
                <NumInput
                    can_cb = { (v, ok) -> self.OnChange(react_obj, "Xp", k, v, ok) }
                    name="Xp"
                    old= { self.GetOne(react_obj.state.v, "Xp", k ) }
                    min="0"
                    max="999999" />
              </div>
        return re

    ModifyRender : (react_obj) ->
        self = @
        <div>
            {self.mkAvatarMods(react_obj)}
        </div>
}
###################################################################


###################################################################
###
{"Level":12,"Xp":43}
###
GameCorpInfo = React.createClass
  getInitialState: () ->
    o = JSON.parse @props.data
    return {
      exps: o
    }

  SetData : (data) ->
    o = JSON.parse data
    @setState {
      exps: o
    }

  render:() ->
      <Table striped bordered condensed hover>
        <thead>
          <tr>
            <th>Name</th>
            <th>Value</th>
          </tr>
        </thead>
        <tbody>
          <tr>
              <td>战队等级</td>
              <td>{@state.exps.Level}</td>
          </tr>
          <tr>
              <td>战队当前经验</td>
              <td>{@state.exps.Xp}</td>
          </tr>
        </tbody>
      </Table>

GameCorpModifyModal = mkModifyModel {
    Get : (v, typ) ->
      obj = JSON.parse v
      return obj[typ]

    OnChange : (react_obj, typ, v, ok) ->
      nv = JSON.parse react_obj.state.v
      nv[typ] = v
      react_obj.setState {
        is_ok : ok
        v : JSON.stringify nv
      }

    mkMods : (react_obj) ->
        self = @
        data = self.Get react_obj.state.v
        return <div>
          <div>Mod Corp Exp: </div>
          <hr />
          <NumInput
              can_cb = { (v, ok) -> self.OnChange(react_obj, "Level", v, ok) }
              name="Level"
              old= { self.Get(react_obj.state.v, "Level" ) }
              min="1"
              max="100" />
          <NumInput
              can_cb = { (v, ok) -> self.OnChange(react_obj, "Xp", v, ok) }
              name="Xp"
              old= { self.Get(react_obj.state.v, "Xp" ) }
              min="0"
              max="999999" />
        </div>

    ModifyRender : (react_obj) ->
        self = @
        <div>
            {self.mkMods(react_obj)}
        </div>
}

###################################################################

###################################################################
###
{"Value":240,"Last_time":1433832902}
###
GameEnergyInfo = React.createClass
  getInitialState: () ->
    o = JSON.parse @props.data
    return {
      data: o
    }

  SetData : (data) ->
    o = JSON.parse data
    @setState {
      data: o
    }

  render:() ->
      <div>
        {@state.data.Value}
      </div>

GameEnergyModifyModal = mkModifyModel {
    Get : (v) ->
      obj = JSON.parse v
      return obj.Value

    OnChange : (react_obj, typ, v, ok) ->
      nv = JSON.parse react_obj.state.v
      nv.Value = v
      react_obj.setState {
        is_ok : ok
        v : JSON.stringify nv
      }

    ModifyRender : (react_obj) ->
        self = @
        <NumInput ref = "in"
            can_cb = { (v, ok) -> self.OnChange(react_obj, "energy", v, ok) }
            name={ react_obj.props.maker.name }
            old= { @Get(react_obj.state.v) }
            min="0"
            max="1000" />
}

###################################################################

###################################################################
###
{
  "Stages": [
    {
      "Id": "k1_2",
      "Reward_state": [
        {
          "Reward_count": 0,
          "Space_num": 0
        },
        {
          "Reward_count": 1,
          "Space_num": 0
        }
      ],
      "T_count": 2,
      "T_refresh": 0,
      "Max_star": 2
    }
  ],
  "Last_update": "2015-06-09T11:28:56.55249256+08:00",
  "LastStageId": "k1_qz1"
}###
GameStageInfo = React.createClass
  getInitialState: () ->
    o = JSON.parse @props.data
    return {
      data: o
    }

  SetData : (data) ->
    o = JSON.parse data
    @setState {
      data: o
    }

  getAllInfo : (data) ->
    if not data?
      return <div>UnKnown Info</div>

    re = []
    for k, v of data.Stages
      console.log v
      re.push <tr key={v.Id} >
              <td>{v.Id}</td>
              <td>{v.T_count}</td>
              <td>{v.T_refresh}</td>
              <td>{v.Max_star}</td>
          </tr>
    return re

  render:() ->
      <div>
        <div>最近一次进入的关卡 {@state.data.LastStageId}</div>
        <Table striped bordered condensed hover>
          <thead>
            <tr>
              <th>关卡ID</th>
              <th>今日通关次数(不会主动刷新)</th>
              <th>今日通关次数刷新次数(不会主动刷新)</th>
              <th>最高星级</th>
            </tr>
          </thead>
          <tbody>
            { @getAllInfo(@state.data) }
          </tbody>
        </Table>
      </div>

GameStageModifyModal = mkModifyModel {
    Get : (v, typ) ->
      obj = JSON.parse v
      return obj[typ]

    OnChange : (react_obj, typ, v, ok) ->
      nv = JSON.parse react_obj.state.v
      nv[typ] = v
      react_obj.setState {
        is_ok : ok
        v : JSON.stringify nv
      }

    mkMods : (react_obj) ->
        self = @
        data = self.Get react_obj.state.v
        return <div>
          <div>Mod Corp Exp: </div>
          <hr />
          <NumInput
              can_cb = { (v, ok) -> self.OnChange(react_obj, "Level", v, ok) }
              name="Level"
               old= { self.Get(react_obj.state.v, "Level" ) }
              min="1"
              max="100" />
           <NumInput
               can_cb = { (v, ok) -> self.OnChange(react_obj, "Xp", v, ok) }
               name="Xp"
               old= { self.Get(react_obj.state.v, "Xp" ) }
               min="0"
               max="999999" />
        </div>

    ModifyRender : (react_obj) ->
        self = @
        <div>
            {self.mkMods(react_obj)}
        </div>
}
###################################################################

###################################################################
###
{"Currency":[123700,70000,357160,0,0,0]}
###
GameCurrencyInfo = React.createClass
  getInitialState: () ->
    o = JSON.parse @props.data
    return {
      data: o.Currency
    }

  SetData : (data) ->
    o = JSON.parse data
    @setState {
      data: o.Currency
    }

  getAllInfo : (data) ->
    if not data?
      return <div>UnKnown Info</div>

    re = []
    for k, v of data
      console.log v
      re.push <tr key={k} >
              <td>{getScName(k)}</td>
              <td>{v}</td>
          </tr>
    return re

  render:() ->
        <Table striped bordered condensed hover>
          <thead>
            <tr>
              <th>类型</th>
              <th>金额</th>
            </tr>
          </thead>
          <tbody>
            { @getAllInfo(@state.data) }
          </tbody>
        </Table>

GameSCModifyModal = mkModifyModel {
    Get : (v) ->
      obj = JSON.parse v
      return obj.Currency

    GetOne : (v, typ) ->
      obj = JSON.parse v
      return obj.Currency[typ]

    OnChange : (react_obj, typ, v, ok) ->
      nv = JSON.parse react_obj.state.v
      nv.Currency[typ] = v
      react_obj.setState {
        is_ok : ok
        v : JSON.stringify nv
      }

    mkMods : (react_obj) ->
        self = @
        re = []
        scs = self.Get react_obj.state.v
        for k, sc of scs
          do (k, sc) ->
            re.push <div>
                <NumInput
                    can_cb = { (v, ok) -> self.OnChange(react_obj, k, v, ok) }
                    name={getScName(k)}
                    old= { self.GetOne(react_obj.state.v, k ) }
                    min="0"
                    max="9999999" />
              </div>
        return re

    ModifyRender : (react_obj) ->
        self = @
        <div>
            {self.mkMods(react_obj)}
        </div>
}
###################################################################

###################################################################
###
{"Currency":[1800,100,0,1900]}
###
GameHardCurrencyInfo = React.createClass
  getInitialState: () ->
    o = JSON.parse @props.data
    return {
      data: o.Currency
    }

  SetData : (data) ->
    o = JSON.parse data
    @setState {
      data: o.Currency
    }

  getAllInfo : (data) ->
    if not data?
      return <div>UnKnown Info</div>

    re = []
    for k, v of data
      console.log v
      re.push <tr key={k} >
              <td>{getHcName(k)}</td>
              <td>{v}</td>
          </tr>
    return re

  render:() ->
        <Table striped bordered condensed hover>
          <thead>
            <tr>
              <th>类型</th>
              <th>金额</th>
            </tr>
          </thead>
          <tbody>
            { @getAllInfo(@state.data) }
          </tbody>
        </Table>

GameHCModifyModal = mkModifyModel {
    Get : (v) ->
      obj = JSON.parse v
      return obj.Currency

    GetOne : (v, typ) ->
      obj = JSON.parse v
      return obj.Currency[typ]

    OnChange : (react_obj, typ, v, ok) ->
      nv = JSON.parse react_obj.state.v
      nv.Currency[typ] = v
      nv.Currency[3] =  nv.Currency[0] +  nv.Currency[1] +  nv.Currency[2]
      react_obj.setState {
        is_ok : ok
        v : JSON.stringify nv
      }

    mkMods : (react_obj) ->
        self = @
        re = []
        scs = self.Get react_obj.state.v
        for k, sc of scs
          do (k, sc) ->
            re.push <div>
                <NumInput
                    can_cb = { (v, ok) -> self.OnChange(react_obj, k, v, ok) }
                    name={getHcName(k)}
                    old= { self.GetOne(react_obj.state.v, k ) }
                    min="0"
                    max="9999999" />
              </div>
        return re

    ModifyRender : (react_obj) ->
        self = @
        <div>
            {self.mkMods(react_obj)}
        </div>
}
###################################################################

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

InfoMakers = {}

RegIM = (key, name, func, mod_func) ->
  InfoMakers[key] = {
    key: key
    name: name
    func: func
    mod_func : mod_func
  }

RegIM "sc",          "软通信息",      ((data, name) -> <GameCurrencyInfo ref={name} data={data} />)
    , (maker, v, cb) -> <GameSCModifyModal maker={maker} v={v} cb={cb} />

RegIM "hc",          "硬通信息",      ((data, name) -> <GameHardCurrencyInfo ref={name} data={data} />)
    , (maker, v, cb) -> <GameHCModifyModal maker={maker} v={v} cb={cb} />

RegIM "energy",      "战队体力",      ((data, name) -> <GameEnergyInfo ref={name} data={data} /> )
    , (maker, v, cb) -> <GameEnergyModifyModal maker={maker} v={v} cb={cb} />

ProfileInfo = React.createClass
  getInitialState: () ->
    {
      profile_data: {}
    }

  SetData : (data) ->
    @setState {
      profile_data : {}
    }
    @setState {
      profile_data : data
    }

  getAllInfo : (data) ->
    self = @
    if not data?
      return <div>Wait To Query</div>

    re = []
    for k, v of data
      do(k,v) ->
        re.push self.getInfoMaker k, v
    return re


  mkMaker : (f, k) ->
    console.log "mkMaker"
    console.log f
    console.log k
    return f(@state.profile_data[k], k)

  getInfoMaker : (k, v) ->
    maker = InfoMakers[k]
    if maker?
      return <tr key={k} >
              <td>{maker.name}</td>
              <td>{@mkMaker(maker.func, k)}</td>
      </tr>
    # 不显示其他的
    else
      return <tr key={k} >
              <td>{k}</td>
              <td>{v}</td>
      </tr>

  render:() ->
      <Table striped bordered condensed hover>
        <thead>
          <tr>
            <th>Class</th>
            <th>Data</th>
          </tr>
        </thead>
        <tbody>
          { @getAllInfo(@state.profile_data) }
        </tbody>
      </Table>

ProfileModify = React.createClass
  getInitialState: () ->
    {
      profile_data: {}
    }

  SetData : (data) ->
    @setState {
      profile_data : {}
    }
    @setState {
      profile_data : data
    }

  GetData : () -> @state.profile_data

  SetDataByKey : (key, data) ->
    console.log "SetDataByKey"
    console.log key
    console.log data
    dataall = @state.profile_data
    @setState {
      profile_data : {}
    }
    dataall[key] = data
    @setState {
      profile_data : dataall
    }

    info = @refs[key]
    if info?
      console.log "data"
      info.SetData data

  getAllInfo : (data) ->
    console.log @state.profile_data
    self = @
    if not data?
      return <div>Wait To Query</div>

    re = []
    for k, v of data
      do(k,v) ->
        re.push self.getInfoMaker k, v
    return re

  getModifyButtons : (maker, v) ->
    console.log("dasdasdasdasdasd ", maker.mod_func, v)
    if maker.mod_func?
      console.log("dasdasdasdasdasd");
      return <div class="container-fluid">
                  <div>{maker.name}</div>
                  <ModalTrigger modal={maker.mod_func(maker, v, @SetDataByKey )}>
                    <Button bsStyle='danger' onClick=null >Modify</Button>
                  </ModalTrigger>
      </div>
    else
      return <div>{maker.name}</div>

  mkMaker : (f, k) ->
    console.log "mkMaker"
    console.log f
    console.log k
    return f(@state.profile_data[k], k)

  getInfoMaker : (k, v) ->
    maker = InfoMakers[k]
    v = {} if not v?
    if maker?
      return <tr key={k} >
              <td>
                {@getModifyButtons maker, v}
              </td>
              <td>{@mkMaker(maker.func, k)}</td>
      </tr>
    # 不显示其他的
    else
      return <tr key={k} >
              <td>{k}</td>
              <td>{v}</td>
      </tr>

  handleChangeDataForCommit : (event) ->
    commit_data = JSON.parse(event.target.value)
    console.log commit_data
    @SetData(commit_data)
  getDataForCommit : () ->
    commit_data = @GetData()
    console.log commit_data
    return JSON.stringify commit_data


  render:() ->
      <div>
          <Table striped bordered condensed hover>
            <thead>
              <tr>
                <th>Class</th>
                <th>Data</th>
              </tr>
            </thead>
            <tbody>
              { @getAllInfo(@state.profile_data) }
            </tbody>
          </Table>
          <div>
              <input type="text" value={@getDataForCommit()} onChange={@handleChangeDataForCommit} />
          </div>
      </div>


ProfileQuery = React.createClass
  getInitialState: () ->
    {
      servers:[]
      # null -> querying -> null
      state:null
      is_can_query:false
      profile_key:null
    }

  componentDidMount: () ->
    $.get "../api/v1/server_all", (result) =>
      @setState {
            servers : JSON.parse(result)
      }

  set_can_query: (profile, is_can) ->
        @setState {
          profile_key : profile
          is_can_query: is_can
        }

  query : () ->
    console.log("query")
    console.log(@props.server)
    console.log(@state.profile_key)
    $.get "../api/v1/profile_get/" + @props.server + "/" + @state.profile_key, (result) =>
        rj = JSON.parse(result)
        for k, v of rj
          console.log v
        console.log rj
        @setProfileInfo rj
    return

  mod : () ->
    server = @refs.selector.GetSelectedProfile()
    ndata = {
      v : JSON.stringify(@getProfileInfo())
    }
    console.log(ndata)
    $.ajax {
        url : "../api/v1/profile_get/" + server + "/" + @state.profile_key,
        dataType: "json",
        type: "POST",
        data: JSON.stringify(ndata),
        contentType: "application/json; charset=utf-8",
        complete: (xhr, status) ->
            rc = {
                status: xhr.status
                json: xhr.responseJSON
            }
            console.log rc
    }

  setProfileInfo : (data) ->
    @refs.info.SetData data

  getProfileInfo : () ->
    if @refs.info?
        return @refs.info.GetData()
    else
        return {}

  getHeader : () ->
    if @props.typ is "Query"
      return "Profile Query"
    else
      return "Profile Modify"

  getList : () ->
    if @props.typ is "Query"
       return <ProfileInfo ref="info" />
    else
       return <ProfileModify ref="info" />

  render:() ->
        <div>
            <div>
                <div id="header">
                    <h1>{@getHeader()}</h1>
                </div>
            </div>
            <div>
                <div id="content">
                    <AccountKeyInput can_cb = @set_can_query />
                </div>
                <p>查询玩家个人身上存档信息</p>
                <p>例: [profile:0:10:XXXXXX][pguild:0:10:XXXXXX][friend:0:10:XXXXXX].....</p>
                <Button
                    bsStyle='primary'
                    onClick= @query
                    disabled={ !@state.is_can_query }
                >
                    Query
                </Button>
            </div>
            <div>
                {@getList()}
            </div>
        </div>

#if document.getElementById('ProfileQuery')?
#  React.render <ProfileQuery typ="Query"  />, document.getElementById('ProfileQuery')

#if document.getElementById('ProfileModify')?
#  React.render <ProfileQuery typ="Modify" />, document.getElementById('ProfileModify')

module.exports = {
    ProfileQuery : ProfileQuery
    GameServerList    : GameServerList
    AccountKeyInput   : AccountKeyInput
}
