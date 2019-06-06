antd    = require 'antd'
Api     = require '../api/api_ajax'
React   = require 'react'
ReactBootstrap = require 'react-bootstrap'

Form            = antd.Form
FormItem        = Form.Item
Select          = antd.Select
Option          = Select.Option
Grid            = ReactBootstrap.Grid
Row             = ReactBootstrap.Row
Col             = ReactBootstrap.Col
Button          = antd.Button
Table           = ReactBootstrap.Table
tbody           = ReactBootstrap.tbody
tr              = ReactBootstrap.tr
td              = ReactBootstrap.td
Checkbox        = antd.Checkbox
CheckboxGroup   = Checkbox.Group

App = React.createClass
    getInitialState: () ->
        ss_init = {}
        ss_init["0"] = "新服"
        ss_init["1"] = "流畅"
        ss_init["2"] = "火爆"
        ss_init["3"] = "拥挤"
        ss_init["4"] = "维护"
        ss_init["5"] = "体验"
        r_ss_init = []
        for k, v of ss_init
            r_ss_init[v]=k
        teamAB_init = {}
        teamAB_init["0"] = "A"
        teamAB_init["1"] = "B"
        teamAB_init["2"] = "AB"
        {
            shard_show_state_const:ss_init
            r_shard_show_state_const : r_ss_init
            shard_show_teamAB: teamAB_init
            gids:[]
            curTeamAB:[]
            select_gid:""
            sid_ss:[]
            selected_sid:{}
            op_show_state:"0"
            op_show_teamAB:"0"
        }

    componentDidMount: () ->
        api = new Api()
        api.Typ("getAllGids")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray()
           .Do (result) =>
                console.log result
                res = JSON.parse(result)
                if res.length > 0
                    def = String(res[0])
                @setState {
                    gids:res
                    select_gid:def
                }

    onGidSelectChange : (value) -> 
        console.log value
        @setState {
            select_gid:value
        }
        api = new Api()
        api.Typ("getShardShowStateByGid")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray([value])
           .Do (result) =>
                console.log result
                res = JSON.parse(result)
                @setState {
                    sid_ss:res
                    selected_sid:{}
                }

    selectOption : () ->
        res = []
        for gid in @state.gids
            str = String(gid)
            res.push(<Option value={str}>{str}</Option>)
        return res

    selectSid : (e) ->
        ss = @state.selected_sid
        for value in e
            ss[value] = value
        console.log ss
        @setState {
            selected_sid:ss
        }

    genGroupCheckBox : (ar) ->
        plainOptions = []
        for a in ar
            plainOptions.push(@decodeShowState(a))
        console.log plainOptions
        return <tr>
                    <td>
                        <CheckboxGroup options={plainOptions} onChange={@selectSid} />
                    </td>
               </tr>

    gencheckbox : () ->
        column_count = 5
        res = []
        selected = []
        count = 0
        for s in @state.sid_ss
            count = count + 1
            if count <= column_count
                selected.push(s)
            if count == column_count
                res.push(@genGroupCheckBox(selected))
                selected = []
                count = 0
        if selected.length > 0
            res.push(@genGroupCheckBox(selected))
            console.log res
        return res

    show_state_select_op : () ->
        res = []
        for k,v of @state.shard_show_state_const
            if k == "4" or k == "5"
                continue
            res.push(<Option value={k}>{v}</Option>)
        return res
    show_teamAB_select_op : () ->
        res = []
        for k,v of @state.shard_show_teamAB
            res.push(<Option value={k}>{v}</Option>)
        return res
    onChgShowState : (value) ->
        @setState {
            op_show_state:value
        }
    onChgTeamAB : (value) ->
            @setState {
                op_show_teamAB:value
            }
    onSetShowTeamAB : () ->
         ss_ar = []
         ss_ar.push(@state.select_gid)
         ss_ar.push(@state.shard_show_teamAB[@state.op_show_teamAB])
         console.log @state.selected_sid
         for k,v of @state.selected_sid
            if v?
                ss_ar.push(@encodeShowState(k))
         console.log ss_ar
         if ss_ar.length > 0
            api = new Api()
            api.Typ("setShardsShowTeamAB")
               .ServerID("")
               .AccountID("")
               .Key(@props.curr_key)
               .ParamArray(ss_ar)
               .Do (result) =>
                    console.log result
                    res = JSON.parse(result)
                    @setState {
                        sid_ss:res
                        selected_sid:{}
                    }
    onSetShowState : () ->
        ss_ar = []
        ss_ar.push(@state.select_gid)
        ss_ar.push(@state.op_show_state)
        console.log @state.selected_sid
        for k,v of @state.selected_sid
            if v?
                ss_ar.push(@encodeShowState(k))
        console.log ss_ar
        if ss_ar.length > 0
            api = new Api()
            api.Typ("setShardsShowState")
               .ServerID("")
               .AccountID("")
               .Key(@props.curr_key)
               .ParamArray(ss_ar)
               .Do (result) =>
                    console.log result
                    res = JSON.parse(result)
                    @setState {
                        sid_ss:res
                        selected_sid:{}
                    }


    encodeShowState : (str) ->
        ss = str.split(" ")
        return ss[0] + " " + @state.r_shard_show_state_const[ss[1]] + ss[2]

    decodeShowState : (str) ->
        ss = str.split(" ")
        show_state = @state.shard_show_state_const[ss[1]]
        if !(show_state?)
            show_state = ""
        return ss[0] + " " + show_state + ss[2]

    render:() ->
        <div>
            <p>说明：</p>
            <p>体验和维护 这两个状态这个页面是没有权限修改的</p>
            <br />

            <Form horizontal>
                <FormItem
                  id="select"
                  label="Gid："
                  labelCol={{ span: 2 }}
                  wrapperCol={{ span: 14 }}>
                  <Select id="select" size="large" style={{ width: 120 }} onChange={@onGidSelectChange}>
                    {@selectOption()}
                  </Select>
                </FormItem>
            </Form>
            <div style={{ marginBottom: 24 }}>
            </div>
            <div>
                <Grid>
                    <Row>
                        <Col xs={12} md={8}>
                            <Table responsive>
                                 <tbody>
                                    {@gencheckbox()}
                                 </tbody>
                            </Table>
                        </Col>
                        <Col xs={6} md={4}>
                            <Select id="select" size="large" defaultValue={@state.op_show_state} style={{ width: 120 }} onChange={@onChgShowState}>
                                {@show_state_select_op()}
                            </Select>
                             <Select id="select" size="large" defaultValue={@state.op_show_state} style={{ width: 120 }} onChange={@onChgTeamAB}>
                                {@show_teamAB_select_op()}
                            </Select>
                            <br/>
                            <Button style={{ marginTop: 24 }} type="primary" onClick={@onSetShowState}>设置</Button>
                            <Button style={{ marginTop: 24 }} type="primary" onClick={@onSetShowTeamAB}>设置TeamAB</Button>
                        </Col>
                    </Row>
                </Grid>
            </div>
        </div>

module.exports = App