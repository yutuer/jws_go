
HelloMessage = React.createClass
  render: ()->
    <div>Hello {@props.name}</div>

ButtonToolbar = ReactBootstrap.ButtonToolbar
Button = ReactBootstrap.Button
MenuItem = ReactBootstrap.MenuItem
DropdownButton = ReactBootstrap.DropdownButton
Modal = ReactBootstrap.Modal

Greeting = React.createClass
  getInitialState: () ->
    {greeted: false}

  greet: () ->
      @setState {greeted: true}
      return 

  render:() ->
      response = @state.greeted ? 'Hi' : ''
      <div>
          <div>Hello {this.props.name}</div>
          <span>{response}</span>
          <button onClick={this.greet}>Hi</button>
      </div>

nullfunc = () ->
  return

Header = React.createClass
    render: () ->
            <ButtonToolbar>
                <DropdownButton bsStyle="link" title="default">
                    <MenuItem eventKey="1">active</MenuItem>
                    <MenuItem divider />
                </DropdownButton>
            </ButtonToolbar>


LoadingButton = React.createClass
  getInitialState : () ->
    {isLoading: false}

  handleClick : () ->
    @setState({isLoading: true})
    setTimeout @setIsLoading, 2000

  setIsLoading : () ->
    @setState({isLoading: false})
    React.render <Test01 />, document.getElementById(@props.Ele)

  onbuttonClick : () ->
    if @isLoading then null else @handleClick

  render: () ->
    isLoading = @state.isLoading;
    <Button
        bsStyle='primary'
        disabled={ isLoading }
        onClick={ @onbuttonClick() } >
        { if isLoading then 'Loading...' else 'Loading state'}
    </Button>

React.render <LoadingButton Ele='content3'/>, document.getElementById('content3')


Test01 = React.createClass
  getInitialState: () ->
    {info: "Wait to Loading ..."}
  onbuttonClick : () ->
    React.render <LoadingButton Ele='content3'/>, document.getElementById('content3')

  onbuttonLoading : () ->
    $.ajax
      url: "http://127.0.0.1:3000/",
      cache: false,
      success: (data) =>
        @setState({info: data})
      error: (xhr, status, err) =>
        console.error(err.toString())


  render : () ->
    <div className='static-modal'>
      <Modal title='Modal title'
        backdrop={false}
        animation={false}
        container={document.getElementById('content3')}
        onRequestHide={ nullfunc }>
        <div className='modal-body'>
          {@state.info}
        </div>
        <div className='modal-footer'>
          <Button 
            onClick={ @onbuttonClick } >
            Close
          </Button>
          <Button
              bsStyle='primary'
              onClick={@onbuttonLoading}
          >Load</Button>
        </div>
      </Modal>
    </div>




#React.render <Header />, document.getElementById('content')

#React.render <Greeting name="foDDDo" />, document.getElementById('content2')
#React.render <Test01 />, document.getElementById('content4')
 