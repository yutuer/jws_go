

function StartRealtime(timestamp) {
    StartEpoch(timestamp);
    StartSSE();
}

function StartEpoch(timestamp) {
    var windowSize = 60;
    var height = 200;
    var defaultData = histogram(windowSize, timestamp);

    window.ccuChart = $('#ccuChart').epoch({
        type: 'time.line',
        axes: ['bottom', 'left'],
        height: height,
        historySize: 10,
        data: [
            {
                label: "Live Bots",
                values: defaultData
            },
            {
                label: "CCU",
                values: defaultData
            }
        ]
    });
}

function StartSSE() {
    if (!window.EventSource) {
        alert("EventSource is not enabled in this browser");
        return;
    }
    var source = new EventSource('/stream');
    source.addEventListener('stats', stats, false);
}

function stats(e) {
    var data = parseJSONStats(e.data);
    ccuChart.push(data.ccu);
}

function parseJSONStats(e) {
    var data = jQuery.parseJSON(e);
    var timestamp = data.timestamp;

    var ccu = [
        {time: timestamp, y: data.LiveBots},
        {time: timestamp, y: data.CCU}

    ];

    return {
        ccu: ccu
    }
}

function histogram(windowSize, timestamp) {
    var entries = new Array(windowSize);
    for(var i = 0; i < windowSize; i++) {
        entries[i] = {time: (timestamp-windowSize+i-1), y:0};
    }
    return entries;
}

var entityMap = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': '&quot;',
    "'": '&#39;',
    "/": '&#x2F;'
};

function rowStyle(nick) {
    var classes = ['active', 'success', 'info', 'warning', 'danger'];
    var index = hashCode(nick)%5;
    return classes[index];
}

function hashCode(s){
  return Math.abs(s.split("").reduce(function(a,b){a=((a<<5)-a)+b.charCodeAt(0);return a&a},0));
}

function escapeHtml(string) {
    return String(string).replace(/[&<>"'\/]/g, function (s) {
      return entityMap[s];
    });
}

window.StartRealtime = StartRealtime
