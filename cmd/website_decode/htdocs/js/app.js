$(document).ready(function () {
    const packet = $('#packet');
    const inputForm =$('#input-form');
    const errDiv = $('#errors')
    const results = $('#result')
    const payloads = $('#json-payloads')

    inputForm.submit(function (event) {
        event.preventDefault();
        const packet = $('#packet').val();
        const refLat = $('#refLat').val();
        const refLon = $('#refLon').val();
        $.get("/decode", {'packet': packet, 'refLat': refLat, 'refLon': refLon})
            .done(function (js) {
                if ("" !== js.Err) {
                    errDiv.text(js.Err)
                    errDiv.show()
                } else {
                    errDiv.hide()
                }
                results.html(js.Description)

                payloads.empty()
                js.Payloads.forEach(pl => {
                    const payload = JSON.parse(pl)
                    console.log(payload)
                    delete payload.Updates
                    let pre = $('<pre/>')
                    pre.html(syntaxHighlight(payload));
                    payloads.append(pre)
                })
            })
            .fail(function () {
                errDiv.text("Failed to get data")
                errDiv.show();
            });
    });

    let search = new URLSearchParams(window.location.search)
    const requestedDecode = search.get("q")
    if ("" !== requestedDecode && null != requestedDecode) {
        packet.val(requestedDecode)
        inputForm.submit()
    } else {
        packet.val("*A028009F96887B05FFA000413602;");
    }


    // hook the menu items
    $('li a[data-packet]').on("click", function () {
        packet.val($(this).data('packet'));
        inputForm.submit()
    });


    $('#getLocation').click(function(e) {
        e.preventDefault()
        console.log("Attempting to get location for this user")
        if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(function (pos) {
                $('#refLat').val(pos.coords.latitude)
                $('#refLon').val(pos.coords.longitude)
            });
        }
    })
});
let examplePackets = {
    21: [],
    17: [],
    18: [],
    20: [],
    16: []
//        28: ['*E1999863859533;']
};
let lastRandomNumber = -1;

function setExamplePacket(df) {
    const length = examplePackets[df].length;
    let id = parseInt(Math.random() * length, 10);
    if (length > 1) {
        while (lastRandomNumber === id) {
            id = parseInt(Math.random() * length, 10);
        }
        lastRandomNumber = id;
    } else {
        id = 0;
    }
    const packet = examplePackets[df][id];
    const packetField = $('#packet');
    packetField.val(packet);
    packetField.submit();
}

const examples = $('#examples');
for (let key in examplePackets) {
    if (examplePackets.hasOwnProperty(key)) {
        let link = $('<a class="button">DF' + key + '</a>');
        link.val("DF " + key);
        link.click(function (event) {
            setExamplePacket(key);
            event.preventDefault()
        });
        examples.append(link);
        examples.append('&nbsp;');

    }
}

function syntaxHighlight(json) {
    if (typeof json != 'string') {
        json = JSON.stringify(json, undefined, 2);
    }
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        let cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}
