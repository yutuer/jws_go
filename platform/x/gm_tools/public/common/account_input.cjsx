React      = require 'react'
ReactBootstrap = require 'react-bootstrap'

AccountKeyInput = React.createClass
  getInitialState : () ->
        {
          value: ''
        }

  IsRight : () -> @validationState() is 'success'

  validationState : () ->
        exp=/^(\d+)\:(\d+)\:(.+)$/
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
      <ReactBootstrap.Input
        type='text'
        value={@state.value}
        placeholder='请输入账号ID...'
        help='账号ID格式必须满足 \"X:X:XXX\"'
        bsStyle={@validationState()}
        hasFeedback
        ref='input'
        groupClassName='group-class'
        labelClassName='label-class'
        onChange={@handleChange} />

module.exports = AccountKeyInput
