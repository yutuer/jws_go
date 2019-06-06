React      = require 'react'
antd        = require 'antd'
Select = antd.Select

langMap = {
      "zh-Hans":  "zh-Hans"
      "zh-SG":"zh-SG"
      "zh-HMT":"zh-HMT"
      "en":"en"
}

LanguagePicker = React.createClass
  handleChange : (data) ->
    @props.handleChange(data)

  fillSelect : () ->
    res = []
    for key of langMap
        res.push <Option value={key}>{langMap[key]}</Option>
    return res

  render:() ->
      <Select defaultValue = "zh-Hans" style={{width:85}} onChange={@handleChange}>
      {@fillSelect()}
      </Select>

MultiLangPicker = React.createClass
    handleChange : (data) ->
        @props.handleChange(data)

    render:() ->
        <Select defaultValue = {@props.defaultValue} style={{width:85}} onChange={@handleChange}>
          <Option value="0">普通</Option>
          <Option value="1">多语言</Option>
        </Select>

MultiLangModal = React.createClass
    getInitialState: () ->
        propsLang = @props.lang
        tempLang = {}
        if !propsLang || !propsLang["zh-Hans"]
            if !@IsJsonString(propsLang)
                for key of langMap
                     tempLang[key] = {title:"", body:""}
            else
                tempLang = JSON.parse(propsLang)
        else
            tempLang = propsLang

        if @props.multi_lang == "1"
            tempTitle = tempLang["zh-Hans"].title
            tempBody = tempLang["zh-Hans"].body
        else
            tempTitle = @props.title
            tempBody = @props.body

        tempMultiLang = @props.multi_lang
        if tempMultiLang != "1"
            tempMultiLang = "0"

        s = {
            title      : tempTitle,
            body       : tempBody,
            language   : tempLang,
            cur_lang   : "zh-Hans",
            multi_lang : tempMultiLang
        }

        return s

    IsJsonString : (jsonStr) ->
        if jsonStr == "" || !jsonStr
            return false
        try
            JSON.parse jsonStr
            return true
        catch err
            return false

    handleChange : (data) ->
        @props.handleChange(data)

    handleTitleChange     : (event) ->
        tempMap = @state.language
        tempMap[@state.cur_lang].title = event.target.value
        @setState { title      : event.target.value, language : tempMap}, ->
            @handleChange(@state)


    handleInfoChange      : (event) ->
        tempMap = @state.language
        tempMap[@state.cur_lang].body = event.target.value
        @setState { body       : event.target.value, language : tempMap}, ->
            @handleChange(@state)

    handleLanguageChange  : (value) ->
        tempMap = @state.language
        @setState {
            cur_lang  : value,
            title      : tempMap[value].title,
            body       : tempMap[value].body
        }, ->
            @handleChange(@state)

    handleMultiLangChange     : (value) ->
        @setState {multi_lang    : value}, ->
            @handleChange(@state)

    render:() ->
        <div>
        <div className="row-flex row-flex-middle row-flex-start">
            <div>多语言:</div>
            <MultiLangPicker defaultValue = {@state.multi_lang} handleChange={@handleMultiLangChange}></MultiLangPicker>
            <div>语言类型:</div>
            <LanguagePicker defaultValue = {@state.cur_lang} handleChange={@handleLanguageChange}></LanguagePicker>
        </div>

        <div className="row-flex row-flex-middle row-flex-start">
            <div>邮件标题:</div>
            <input className="ant-input" value = {@state.title} onChange={@handleTitleChange}></input>
        </div>
        <p>邮件内容:</p>
        <textarea className="ant-input" value = {@state.body} onChange={@handleInfoChange}></textarea>
        </div>

module.exports = {
    LanguagePicker : LanguagePicker
    MultiLangPicker: MultiLangPicker
    langMap : langMap
    MultiLangModal : MultiLangModal
}