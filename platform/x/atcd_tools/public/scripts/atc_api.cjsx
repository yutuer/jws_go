AtcRestClient = (endpoint) ->
    _add_ending_slash = (string) ->
        string += "/"  unless string[string.length - 1] is "/"
        return string

    @endpoint = endpoint or "/api/v1/"
    @endpoint = _add_ending_slash(@endpoint)

    @api_call = (method, urn, callback, data) ->
      urn = _add_ending_slash(urn)
      $.ajax
        url: @endpoint + urn
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


AtcRestClient::shape = (callback, data) ->
  @api_call "POST", "shape", callback, data

AtcRestClient::unshape = (callback, data) ->
  @api_call "DELETE", "shape", callback

AtcRestClient::getCurrentShaping = (callback) ->
  @api_call "GET", "shape", callback

AtcRestClient::getToken = (callback) ->
  @api_call "GET", "token", callback

AtcRestClient::getAuthInfo = (callback) ->
  @api_call "GET", "auth", callback

AtcRestClient::updateAuthInfo = (address, data, callback) ->
  @api_call "POST", "auth/".concat(address), callback, data