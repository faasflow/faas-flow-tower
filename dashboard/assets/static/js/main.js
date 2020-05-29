window.chartColors = {
      red: 'rgb(255, 99, 132)',
      orange: 'rgb(255, 159, 64)',
      yellow: 'rgb(255, 205, 86)',
      green: 'rgb(75, 192, 192)',
      blue: 'rgb(54, 162, 235)',
      purple: 'rgb(153, 102, 255)',
      grey: 'rgb(201, 203, 207)'
};


// Get the server address
function getServer() {
    let server = "";
    if (serverAddr != "") {
        server = serverAddr;
    } else {
        let protocol = location.protocol;
        let slashes = protocol.concat("//");
        server = slashes.concat(window.location.hostname);
    }
    return server;
};

// d3 dot vix attributor
function attributer(datum, index, nodes) {
    let selection = d3.select(this);
    if (datum.tag == "svg") {
        let width = window.innerWidth * 0.6;
        let height = window.innerHeight * 0.5;
        let x = 0;
        let y = 0;
        let scale = 0.3;
        selection
            .attr("width", width + "pt")
            .attr("height", height + "pt")
            .attr("viewBox", -x + " " + -y + " " + (width / scale) + " " + (height / scale));
        datum.attributes.width = width + "pt";
        datum.attributes.height = height + "pt";
        datum.attributes.viewBox = -x + " " + -y + " " + (width / scale) + " " + (height / scale);
    }
};

// updateGraph updates the graph
function updateGraph(dag) {
    d3.select("#graph")
        .graphviz()
        .tweenShapes(false)
        .attributer(attributer)
        .renderDot(dag);
};

// Load the trace content async and periodic manner
function loadTraceContent(traceId) {
    let url = getServer();
    url = url.concat("/function/faas-flow-dashboard/api/flow/request/traces");

    let reqData = {};
    reqData["method"] = "request-traces";
    reqData["trace-id"] = traceId;

    let data = JSON.stringify(reqData);

    let xmlHttp = new XMLHttpRequest();
    xmlHttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status != 200) {
            alert("Failed to get flow details from : " + url);
            return;
        }
        if (this.readyState == 4 && this.status == 200) {
            updateTraceContent(JSON.parse(this.responseText));
        }
    };
    xmlHttp.open("POST", url, true);
    xmlHttp.setRequestHeader('accept', "application/json");
    xmlHttp.setRequestHeader("Content-Type", "application/json");
    xmlHttp.send(data);
};

// delete the flow function
function deleteFlow(flowName) {
    let url = getServer();
    url = url.concat("/function/faas-flow-dashboard/api/flow/delete");

    let reqData = {};
    reqData["function"] = flowName;
    let data = JSON.stringify(reqData);

    let xmlHttp = new XMLHttpRequest();
    xmlHttp.onreadystatechange = function () {
        if (this.readyState == 4 && this.status != 200) {
            alert("Failed to delete flow: " + flowName);
            return;
        }
        if (this.readyState == 4 && this.status == 200) {
            alert("Deleted flow: " + flowName);
            return;
        }
    };
    xmlHttp.open("POST", url, true);
    xmlHttp.setRequestHeader('accept', "application/json");
    xmlHttp.setRequestHeader("Content-Type", "application/json");
    xmlHttp.send(data);
};

// format function duration in sec
function formatDuration(micros) {
    let seconds = (micros / 1000000);
    return "" + seconds + "s";
};

// format Time in hour:min:sec
function formatTime(unix_timestamp) {
    let date = new Date(unix_timestamp/1000);
    let hours = date.getHours();
    let minutes = "0" + date.getMinutes();
    let seconds = "0" + date.getSeconds();

    // Will display time in 10:30:23 format
    let formattedTime = hours + ':' + minutes.substr(-2) + ':' + seconds.substr(-2);
    return formattedTime
};

// draw the bar chart for request tarces
function drawBarChart(jsonObject) {
    let id = jsonObject["request-id"];
    let rstime = jsonObject["start-time"];
    let rduration = jsonObject["duration"];
    let traces = jsonObject["traces"];
  
    let container = document.getElementById('canvas');
    let chart = new google.visualization.Timeline(container);
    let dataTable = new google.visualization.DataTable();


    dataTable.addColumn({ type: 'string', id: 'ID' });
    dataTable.addColumn({ type: 'number', id: 'Start' });
    dataTable.addColumn({ type: 'number', id: 'End' });
    
    let rows = [];

    let normalizer = 1000;
    let requestdata = [id, (rstime/normalizer), ((rstime+rduration)/normalizer)];
    rows.push(requestdata);

    for (let node in traces) {
        let value = traces[node];
        let nstime = value["start-time"];
        let nduration = value["duration"];
        rows.push([node, nstime/normalizer, ((nstime+nduration)/normalizer)]);
    }
    dataTable.addRows(rows)

    let options = {
        animation:{
            duration: 1000,
            easing: 'out',
        },
	    hAxis:{
	        minValue: (rstime/normalizer),
	        maxValue: ((rstime+rduration)/normalizer),
	    },
    };
    chart.draw(dataTable, options);
};

// Update the content of content wrapper for request desc
function updateTraceContent(jsonObject) {

    let duration = jsonObject["duration"];
    let status = jsonObject["status"];
    let start_time = jsonObject["start-time"];

    // remove welcome body if present
    welcome = d3.select("#welcome")
    if (welcome !== null ) {
        welcome.remove();
    }

    // draw the bar chart
    google.charts.load("current", {
	    packages: ["timeline"]},
    );
    google.charts.setOnLoadCallback(function(){
	    drawBarChart(jsonObject);
    });

    d3.select("#exec-duration").text("Duration: " + formatDuration(duration));
    d3.select("exec-status").text("Status: " + status);
    d3.select("start-time").text("Start Time: " + formatTime(start_time));
};



