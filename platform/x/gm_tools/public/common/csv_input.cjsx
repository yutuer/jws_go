boot  = require 'react-bootstrap'
React = require 'react'
Input = boot.Input
Table = boot.Table

CSVInput = React.createClass
    getInitialState : () ->
        {
            texts : ""
            values: []
        }

    componentWillReceiveProps : (nextProps) ->
        console.log "componentWillReceiveProps " + nextProps.title
        if @state.texts != nextProps.value
            res = @mkValues nextProps.value
            @setState {
                texts: nextProps.value
                values: res
            }

    IsRight : () ->
        res = true
        for idx, v of @state.values
            res = res and (@validationState(v) is 'success')
        return res

    validationState : ( value ) ->
        if value?
            return 'success'
        else
            return 'error'

    handleChange : () ->
        console.log "handleChange"
        v = @refs.input.getValue()
        res = @mkValues v
        console.log res
        @setState {
            values: res
        }
        console.log v
        @setState {
            texts: v
        }
        @props.can_cb(res, v, @IsRight())

    mkValues : (texts) ->
        t = texts.split '\n'
        return t.map (v) -> v.split ';'

    
    getTable : () ->
        <Table  {...@props} striped bordered condensed hover >
            <tbody>
                {@state.values.map (v) ->
                    <tr key={v.join("-")} >
                        {v.map (t) -> <td>{t}</td>}
                    </tr>}
            </tbody>
        </Table>

    render : () ->
        <div className="row-flex row-flex-start">
            <Input
                ref         = "input"
                type        = "textarea"
                label       = {@props.title}
                value       = {@state.texts}
                onChange    = {@handleChange} />
            {@getTable()}
        </div>


module.exports = CSVInput
