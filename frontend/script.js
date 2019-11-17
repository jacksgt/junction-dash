function getSensors() {
    return {"ant": "Ant", "bear": "Bear", "cheetah": "Cheetah", "dolphin": "Dolphin"};
}

function openSocket(mac) {
    socket = new WebSocket("ws://u9k.de/mac/" + encodeURIComponent(mac));
    socket.onmessage = onMessage;
    return socket;
}

function onMessage(event) {
    data = JSON.parse(event.data);
    console.log(data);
    if (data["type"] === "update") {
        $.each(data["values"], function (sensor, value) {
            setNode(sensor, value);
        });
        blinkLed();
    } else if (data["type"] === "complete") {
        completeSensor(data["station"]);
    }

}

function blinkLed() {
    console.log("Activate led..");
    $("#activityLed").addClass("on");
    setTimeout(function() {
        console.log("Turn off led..");
        $("#activityLed").removeClass("on");
    }, 500);
}

$(document).ready(function () {
    $.ajax("http://u9k.de/sensor/").done(function (body) {
        const data = JSON.parse(body);
        $.each(data, function (sensor, name) {
            var node ="<div style='padding-top:10px;text-align: center;'><div id=\"node-" + sensor + "\" class=\"circle\">" + name + "</div>";
            node += "<button id=\"sound-" + sensor + "\" type=\"submit\" class=\"sound btn btn-sm btn-secondary\">Play Sound</button></div>";
    
            $("#nodeList").append(node)

            $("#sound-" + sensor).click(function() {
                $(this).attr("disabled", true);
                
                if (localStorage.getItem(sensor) !== "set") {
                    localStorage.setItem(sensor, "set");
                    $.post("http://u9k.de/sound/", {station: sensor}, function(data, status) {
                        console.log(data + sensor);
                        $("#sound-" + sensor).text(data);
                    });
                } else {
                    $(this).text("Already pressed");
                }
            })
        });
        $("#macForm").submit(function (event) {
            const mac = $("#macInput").val();
            openSocket(mac);
            event.preventDefault();
            $("#submitMAC").attr("disabled", true);

        });
    });
});

function setNode(sensor, value) {
    const id = "#node-" + sensor;
    const hue = Math.max(0, Math.min(120, 120 + 0.45 * (value + 50)));
    const color = "hsl(" + hue + ", 80%, 50%)";
    $(id + ":not(.completed)").css("background", color);
}

function completeSensor(sensor) {
    $("#node-" + sensor)
        .addClass("completed");
    if(!triggeredAlert){
        notifyDanger("Looks like you found "+sensor+"!")
        triggeredAlert = 1

    }
}
var triggeredAlert = 0
function notifyDanger(message){
    // Set alert message, show, then hide after 10 seconds
    $("#notification").text(message.toString());
    $("#notification").show();
    setTimeout(function() {
        $("#notification").hide();
    }, 10000);
    

}
function notifyNormal(message){
    // Set alert message, show, then hide after 10 seconds
    $("#notification").text(message.toString());
    $("#notification").show();
    setTimeout(function() {
        $("#notification").hide();
    }, 10000);
    

}

