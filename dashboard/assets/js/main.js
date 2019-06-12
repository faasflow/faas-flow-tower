// Get the server address
function getServer() {
    var server = "";
    if (serverAddr != "") {
        server = serverAddr;
    } else {
        var protocol = location.protocol;
        var slashes = protocol.concat("//");
        server = slashes.concat(window.location.hostname);
    }
    return server;
};


function attributer(datum, index, nodes) {
    var selection = d3.select(this);
    if (datum.tag == "svg") {
        var width = window.innerWidth * 0.6;
        var height = window.innerHeight * 0.5;
        var x = 0;
        var y = 0;
        var scale = 0.3;
        selection
            .attr("width", width + "pt")
            .attr("height", height + "pt")
            .attr("viewBox", -x + " " + -y + " " + (width / scale) + " " + (height / scale));
        datum.attributes.width = width + "pt";
        datum.attributes.height = height + "pt";
        datum.attributes.viewBox = -x + " " + -y + " " + (width / scale) + " " + (height / scale);
    }
};

// Update the content of content wrapper
function updateContent(jsonObject) {
    var name = jsonObject["name"];
    var count = jsonObject["invocationCount"];
    var replicas = jsonObject["replicas"];
    var dag = jsonObject["dag"];
    var description = "No flow description provided. Use faas-flow-desc in lebels";
    if ("faas-flow-desc" in jsonObject["labels"]) {
	description = jsonObject["labels"]["faas-flow-desc"];
    }
   
    var url = getServer();
    url = url.concat("/function/" + name);

    // Remove welcome body if present
    d3.select("#welcome").remove();

    d3.select("#flow-name").text(name);
    d3.select("#flow-desc").text(description);
    d3.select("#exec-count").text("Execution Count: " + count);
    d3.select("#replica-count").text("Replicas: " + replicas);
    d3.select("#link").attr("href", url);
    // remove and render new graph
    d3.select("#graph")
     .selectAll("*")
     .remove();

    d3.select("#page-content-wrapper").style("visibility", "visible");
    d3.select("#graph")
     .graphviz()
     .tweenShapes(false)
     .attributer(attributer)
     .renderDot(dag);
};

// Add event listener to all faas-flow function
document.getElementsByName("function-switch").forEach(function(elem) {
    elem.addEventListener("click", function(event) {
         var functionId = elem.getAttribute("id");
         var url = getServer();
         url = url.concat("/function/faas-flow-dashboard");

         var reqData = {};
         reqData["method"] = "flow";
         reqData["function"] = functionId;

         data = JSON.stringify(reqData);

         var xmlHttp = new XMLHttpRequest();
         xmlHttp.onreadystatechange = function() {
             if (this.readyState == 4 && this.status != 200) {
                alert("Failed to get flow details from : " + url);
                return;
             }
             if (this.readyState == 4 && this.status == 200) {
               var jsObj = JSON.parse(this.responseText);
               updateContent(jsObj);
             }
         };
         xmlHttp.open("POST", url, true);
         xmlHttp.setRequestHeader('accept', "application/json");
         xmlHttp.setRequestHeader("Content-Type", "application/json");
         xmlHttp.send(data);
    });
});

