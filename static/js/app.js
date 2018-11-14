/*
 * This is the code for your app that will run in the *browser* at the client side.
 */

// This will run once when the browser has finished retrieving all the static files and the page has finished loading.
$(function () {

    // Create page content.
    var startContent = "<h1>My Web App</h1>";
    startContent += "<h3>My Number: <span id='numID'>0</span></h3>";
    startContent += "<button id='btnID' class='btn btn-default' type='submit'>Increment Number</button>";

    // Insert content into page (DOM).
    $("#UIcontent").html(startContent);

    // Run provided function when button is clicked.
    var numDOMhandle = $("#numID"); // Get a handle to the number shown on the page.
    $("#btnID").click(function () {
        var queryParameters = { number: numDOMhandle.html() }; // Set number as query parameter for request.
        $.get("changenumber", queryParameters, function (response) { // HTTP request to back-end, sending the number.
            numDOMhandle.html(response); // On request success, update the number on the page.
        });
    });

    // Setup websocket connection
    conn = new WebSocket("ws://" + document.location.host + "/ws");
    conn.onclose = function () { console.error("ws conn closed"); };
    var wsNumHandle = $("#shared-num");
    conn.onmessage = function (msg) {
        wsNumHandle.html(msg.data);
    };
    $("#shared-num-inc").click(function () {
        conn.send(wsNumHandle.html());
    });
});
