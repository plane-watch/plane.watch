$(document).ready(function () {
    let packet = $('#packet');
    let inputForm =$('#input-form');
    inputForm.submit(function (event) {
        event.preventDefault();
        let packet = $('#packet').val();
        let refLat = $('#refLat').val();
        let refLon = $('#refLon').val();
        $.get("/decode", {'packet': packet, 'refLat': refLat, 'refLon': refLon})
            .done(function (data) {
                $('#result').html(data)
            })
            .fail(function () {
                $('#result').html("Failed to get data")
            });
    });

    let search = new URLSearchParams(window.location.search)
    const requestedDecode = search.get("q")
    if ("" !== requestedDecode) {
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
        var link = $('<a class="button">DF' + key + '</a>');
        link.val("DF " + key);
        link.click(function (event) {
            setExamplePacket(key);
            event.preventDefault()
        });
        examples.append(link);
        examples.append('&nbsp;');

    }
}
