Api         = require '../api/api_ajax'
React      = require 'react'
class SysRollNotice
    constructor: ( @json ) ->
        @begin     = @json.begin
        @end       = @json.end
        @id        = @json.id
        @interval  = @json.interval
        @state     = @json.command.params[3]
        @title     = @json.command.params[4]
        @info      = @json.command.params[5]
        @multi_lang = @json.command.params[7]
        @language = @json.command.params[8]
        @server_id = @json.command.server
        [ @gid, @sid ] = @server_id.split ':'

    Log   : () -> console.log(@)
    State : () ->
        return "未发布" if @state is "0"
        now_t = new Date()
        begin_t = new Date(@json.command.params[0])
        end_t = new Date(@json.command.params[1])
        return "已发布" if begin_t <= now_t < end_t
        return "已过期" if now_t >= end_t
        return "未到期" if now_t < begin_t

    ChangeState : () ->
        console.log @state
        if @state is "0"
            @state = "1"
        else
            @state = "0"


    UpdateToServer : (server_id, account_id, curr_key, cb) ->
        console.log @
        console.log "curr_key"
        console.log curr_key
        api = new Api()

        if @begin >= 10000000000
            start = new Date(@begin)
            hours = (start.getHours()).toString()
        else
            start = new Date(@begin*1000)
            hours = (start.getHours()-1).toString()
        year = start.getFullYear()
        month = (start.getMonth() + 1).toString()
        day = start.getDate().toString()

        minutes = start.getMinutes().toString()
        seconds = start.getSeconds().toString()
        if month.length == 1
         month = '0' + month
        if day.length == 1
         day = '0' + day
        if hours.length == 1
          hours = '0' + hours
        if minutes.length == 1
          minutes = '0' + minutes
        if seconds.length == 1
          seconds='0'+ seconds
        start_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" +seconds

        if @end >= 10000000000
            end = new Date(@end)
            hours = (end.getHours()).toString()
        else
            end = new Date(@end*1000)
            hours = (end.getHours()-1).toString()
        console.log end
        year = end.getFullYear()
        month = (end.getMonth() + 1).toString()
        day = end.getDate().toString()
        minutes = end.getMinutes().toString()
        seconds = end.getSeconds().toString()
        if month.length == 1
         month = '0' + month
        if day.length == 1
         day = '0' + day
        if hours.length == 1
          hours = '0' + hours
        if minutes.length == 1
          minutes = '0' + minutes
        if seconds.length == 1
          seconds='0'+ seconds
        end_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" +seconds


        api.Typ("updateSysRollNotice")
           .ServerID(server_id)
           .AccountID(account_id)
           .Key(curr_key)
           .Params(
               start_day_time, 
               end_day_time,
               @interval,
               @state,
               @title,
               @info,
               @id,
               @multi_lang,
               JSON.stringify(@language),
           )
           .Do (result) => cb() if cb?

    NewToServer : (server_id, account_id, curr_key, cb) ->
        api = new Api()

        start = new Date(@begin)
        year = start.getFullYear()
        month = (start.getMonth() + 1).toString()
        day = start.getDate().toString()
        hours = (start.getHours()).toString()
        minutes = start.getMinutes().toString()
        seconds = start.getSeconds().toString()
        if month.length == 1
         month = '0' + month
        if day.length == 1
         day = '0' + day
        if hours.length == 1
          hours = '0' + hours
        if minutes.length == 1
          minutes = '0' + minutes
        if seconds.length == 1
          seconds='0'+ seconds
        start_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" +seconds
        end = new Date(@end)
        year = end.getFullYear()
        month = (end.getMonth() + 1).toString()
        day = end.getDate().toString()
        hours = (end.getHours()).toString()
        minutes = end.getMinutes().toString()
        seconds = end.getSeconds().toString()
        if month.length == 1
         month = '0' + month
        if day.length == 1
         day = '0' + day
        if hours.length == 1
          hours = '0' + hours
        if minutes.length == 1
          minutes = '0' + minutes
        if seconds.length == 1
          seconds='0'+ seconds
        end_day_time = year + "/" + month + "/" + day + " " + hours + ":" + minutes + ":" +seconds


        api.Typ("sendSysRollNotice")
           .ServerID(server_id)
           .AccountID(account_id)
           .Key(curr_key)
           .Params(
               start_day_time, 
               end_day_time,
               @interval,
               @state,
               @title,
               @info,
               @id,
               @multi_lang,
               JSON.stringify(@language)
           )
           .Do (result) => cb() if cb?



module.exports = SysRollNotice
