antd         = require 'antd'
ReactCookie  = require 'react-cookie'
$            = require 'jquery'
notification = antd.notification

openNotificationWithIcon = (type, title, msg) ->
    notification[type] {
        message: title,
        description: msg
        duration: 5
    }

class Api
    Typ : (typ) ->
        @command_typ = typ
        return @

    ServerID : (id) ->
        @server_id = id
        return @

    AccountID : (id) ->
        @account_id = id
        return @

    Key : (id) ->
        @key = id
        return @

    Params : () ->
        if arguments?
            res = []
            res.push p.toString() for p in arguments
            @params = res
        else
            @params = []
        console.log(@params)
        return @

    ParamArray : (a) ->
        @params = a
        console.log(@params)
        return @

    Do : (on_success) ->
        if not @params?
            @params = []
        ndata = {
            "params" : @params,
            "key"    : @key
        }

        console.log(ndata)
        self = @
        $.ajax {
            url : @mkUrl(),
            dataType: "json",
            type: "POST",
            data: JSON.stringify(ndata),
            contentType: "application/json; charset=utf-8",
            complete: (xhr, status) ->
                rc = {
                    status: xhr.status
                    json: xhr.responseJSON
                }
                console.log xhr

                if xhr.status is 200
                    openNotificationWithIcon('success', self.command_typ, xhr.responseText)
                else
                    if xhr.status is 401
                        ReactCookie.save 'taihe-gm-tool-key', "", {
                            path    : '/'
                        }
                    openNotificationWithIcon('error', self.command_typ, xhr.responseText)

                if on_success?
                    on_success(xhr.responseText, xhr.responseJSON, xhr.status)
        }

    ###
        /api/v1/command/0:0/0:0:111/adfadsf/getVirtualIAP
    ###
    mkUrl : () ->
        if not @command_typ?
            console.error "Api mkUrl Err by no @command_typ"
            return null
        if not @server_id?
            console.error "Api mkUrl Err by no @server_id"
            return null
        if not @key?
            console.error "Api mkUrl Err by no @key"
            return null
        if not @account_id? or @account_id is ""
            @account_id = "ac"
        return "../api/v1/command/" + @server_id + "/" + @account_id  + "/" + @command_typ



module.exports = Api
