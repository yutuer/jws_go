antd    = require 'antd'
Api     = require '../api/api_ajax'
React   = require 'react'
ReactBootstrap = require 'react-bootstrap'

Form            = antd.Form
FormItem        = antd.Form.Item
Input           = antd.Input
Button          = antd.Button
Table           = antd.Table
Modal           = antd.Modal
Select          = antd.Select
Option          = Select.Option
Grid            = ReactBootstrap.Grid
Row             = ReactBootstrap.Row
Col             = ReactBootstrap.Col
tbody           = ReactBootstrap.tbody
tr              = ReactBootstrap.tr
td              = ReactBootstrap.td
Checkbox        = antd.Checkbox
CheckboxGroup   = Checkbox.Group

defaultPlainOptions = ['0:10', '0:11', '0:12']

Demo = React.createClass
    getInitialState: () ->
        {
            vershard        :[]
            shardHotInfo    :[]
            batch_version   :""
            select_version  :""
            select_gidsid   :""
            modal_visible   :false
            modal_title     :""
            modal_content   :""
            selected_sid    :[]
            all_sid         :[]
        }

    handleSubmit_ver : (e) ->
        e.preventDefault()
        api = new Api()
        api.Typ("getServerByVersion")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray([@props.form.getFieldValue("version")])
           .Do (result) =>
                console.log "result"+result
                res = JSON.parse(result)
                @onReceiveSidList(res["Infos"])
                @setState {
                   vershard: res["Infos"]
                }

    handleSubmit_data_old : (e) ->
        e.preventDefault()

        sp = [@props.form.getFieldValue("version"), @props.form.getFieldValue("datac")]
        for v in @state.selected_sid
            if v?
                sp.push(v)
        console.log sp

        api = new Api()
        api.Typ("setVersionHotC")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray([@props.form.getFieldValue("version")])
           .Do (result) =>
                console.log "result"+result
                res = JSON.parse(result)
                @setState {
                   vershard: res["Infos"]
                   batch_version: ver
                }

    handleSubmit_data : (e) ->
        e.preventDefault()

        sp = [@props.form.getFieldValue("version"), @props.form.getFieldValue("datac")]
        for v in @state.selected_sid
            if v?
                sp.push(v)
        console.log sp

        ver = @props.form.getFieldValue("version")
        api = new Api()
        api.Typ("setVersionHotC")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray(sp)
           .Do (result) =>
               console.log "result "+result
               @setState {
                   batch_version: ver
               }

    onReceiveSidList: (data) ->
        all_sid = []
        for s in data
            all_sid.push(s)

        @setState {
            all_sid: all_sid
            selected_sid: []

        }

    GetShardDataInfo : () ->
        api = new Api()
        api.Typ("getServerHotData")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray([@props.form.getFieldValue("version")])
           .Do (result) =>
                console.log "result"+result
                res = JSON.parse(result)
                @setState {
                   shardHotInfo: res["Infos"]
                }

    table_shards_columns: () ->
        that = this
        return [{
          title: '版本号',
          dataIndex: 'version'
          sorter : (a, b) ->
            return a.version.localeCompare(b.version)
        },{
          title: '数据热更号',
          dataIndex: 'datac',
          sorter : (a, b) ->
            return a.datac - b.datac
        },{
          title: 'GidSid',
          dataIndex: 'gidsid',
          sorter : (a, b) ->
            return a.shard - b.shard
        },{
          title: '基础数据build',
          dataIndex: 'basebuild',
        },{
          title: '热更数据号',
          dataIndex: 'hotdatac',
        },{
          title: '热更数据build',
          dataIndex: 'hotbuild',
        },{
          title: '发热更信号',
          key: 'signal',
          render : (text, record) ->
            if record.datac != "" and Number(record.datac) > 0
              return <Button type='primary' onClick= {
                      () =>
                            console.log "signal "+record.gidsid
                            that.setState {
                                modal_visible:true
                                select_version:record.version
                                select_gidsid:record.gidsid
                                modal_title:"版本号: "+record.version+"  服务器: "+record.gidsid
                                modal_content: "热更数据: "+record.datac
                            }

                      }  >
                        操作
                     </Button>
            else
              return <div> </div>

        }]

    table_shards_content: () ->
        i = 0
        res = []
        for info in @state.shardHotInfo
            res.push({
                key:i,
                version:info["Version"],
                datac:info["VerSeq"],
                gidsid:info["GidSid"],
                basebuild:info["SidBaseHotBuild"],
                hotdatac:info["SidHotSeq"],
                hotbuild:info["SidHotBuild"],
                })
            i += 1
        return res

    modalOk : () ->
        api = new Api()
        api.Typ("signalServerHotC")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray([@state.select_gidsid, @state.select_version])
           .Do (result) =>
               console.log "result "+result
        @setState {
            modal_visible:false
        }

    modalCancel : () ->
        @setState {
            modal_visible:false
        }

    dataUpdate : () ->
        param = []
        param.push(@state.batch_version)
        for v in @state.selected_sid
           if v?
              param.push(v)

        console.log "param "+param
        api = new Api()
        api.Typ("signalAllServerHotC")
           .ServerID("")
           .AccountID("")
           .Key(@props.curr_key)
           .ParamArray(param)
           .Do (result) =>
               console.log "result "+result

    selectSid : (e) ->
        ss = []
        for value in e
            ss.push(value)
        @setState {
            selected_sid:ss
        }

    onCheckAllChange: (e) ->
        if e.length > 0
            @selectSid(@state.all_sid)
        else
            @setState {
                selected_sid:[]
            }

    genSelectAllCheckBox : () ->
         return <div>
                    <div>
                        <CheckboxGroup options={["全选"]} onChange={@onCheckAllChange} />
                    </div>
               </div>

    genGroupCheckBox : (ar) ->
        plainOptions = []
        for a in ar
            plainOptions.push(a)
        console.log plainOptions
        return <div style={{width: 500}}>
                <CheckboxGroup options={plainOptions} value={@state.selected_sid} onChange={@selectSid}/>
            </div>


    gencheckbox : () ->
        res = []
        res.push(@genGroupCheckBox(@state.all_sid))
        if res.length > 0
            res.unshift(@genSelectAllCheckBox())
        return res

    render:() ->
        { getFieldProps } = @props.form
        <div>
            <p>说明：</p>
            <p>用于服务器数据热更新</p>
            <p>1、输入版本号，[获取指定版本的服务器]</p>
            <p>2、输入要更新的数据版本号，并勾选要更新的服务器，并[提交]</p>
            <p>3、此时可以[获取服务器数据信息]，在要更新的服务器后面点[操作]执行单个更新</p>
            <p>3、此时也可以[热更新]，更新全部选中的服务器</p>
            <br />
            <Form inline onSubmit={@handleSubmit_ver}>
                <FormItem label="版本号: ">
                  <Input placeholder=""
                    {...getFieldProps('version')}
                  />
                </FormItem>

                <Button type="primary" htmlType="submit">获取指定版本的服务器</Button>
            </Form>
                     

            <Form inline onSubmit={@handleSubmit_data}>

                <FormItem label="数据No.: ">
                  <Input placeholder="请输入数据打包号"
                    {...getFieldProps('datac')}
                  />
                </FormItem>
                <Button type="primary" htmlType="submit">提交</Button>
            </Form>

            <Button style={{ marginTop: 24 }} type="primary" onClick={@dataUpdate}>热更新</Button>

            <div style={{ marginBottom: 24 }}>
            <br />
            </div>

            <div>
                {@gencheckbox()}
            </div>


            <div style={{ marginBottom: 24 }}>
            <br />
            </div>
            

            <Form horizontal>
                <FormItem>
                    <Button type="primary" onClick={@GetShardDataInfo}>获取服务器数据信息</Button>
                </FormItem>
                <FormItem>
                    <Table columns={@table_shards_columns()} dataSource={@table_shards_content()} />
                </FormItem>
            </Form>

            <Modal title={@state.modal_title} visible={@state.modal_visible}
              onOk={@modalOk} onCancel={@modalCancel}>
              是否确定{@state.modal_content}
            </Modal>
        </div>

App = Form.create()(Demo)
module.exports = App