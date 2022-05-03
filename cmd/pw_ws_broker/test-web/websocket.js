;(() => {
    console.log("Websockets!")

    let host = window.location.host

    let ws = {
        conn: null,
        dial: function () {
            this.conn = new WebSocket(`ws://${host}/planes`, ["planes"])
            this.conn.addEventListener("close", ev => {
                console.error(`WebSocket Disconnected code: ${ev.code}, reason: ${ev.reason}`)
                if (ev.code !== 1001) {
                    console.info("Reconnecting in 1s")
                    setTimeout(ws.dial, 1000)
                }
            })
            this.conn.addEventListener("open", ev => {
                console.info("websocket connected")
            })

            // This is where we handle messages received.
            this.conn.addEventListener("message", ev => {
                if (typeof ev.data !== "string") {
                    console.error("unexpected message type", typeof ev.data)
                    return
                }

                const payload = JSON.parse(ev.data)

                if ('type' in payload) {
                    console.info(payload.type)
                    switch (payload.type) {
                        case 'error':
                            console.error("API Error, message=", payload.message)
                            break;
                        case 'plane-location':
                            break;
                        case 'plane-location-list':
                            if ('locations' in payload) {
                                console.info(payload.locations.length)
                                if (payload.locations.length > 0) {
                                    console.info(payload.locations[0])
                                }
                            }
                            break;

                    }
                }
            })
        },
        tiles: function() {
            this.send({
                type: "sub-list",
            })
        },
        sub: function(tile) {
            this.send({
                type: "sub",
                gridTile: tile,
            })
        },
        unsub: function(tile) {
            this.send({
                type: "unsub",
                gridTile: tile,
            })
        },
        currentPlanes: function (tile) {
            this.send({
                type: "grid-planes",
                gridTile: tile
            })
        },
        planeHistory(icao, callSign) {
            this.send({
                type: 'plane-location-history',
                icao: icao,
                callSign: callSign,
            })
        },
        send: function(msg) {
            this.conn.send(JSON.stringify(msg))
        }
    }
    ws.dial()

    function getJSON(url, callback) {
        const xhr = new XMLHttpRequest();
        xhr.open('GET', url, true);
        xhr.responseType = 'json';
        xhr.onload = function () {
            const status = xhr.status;
            if (status === 200) {
                callback(xhr.response);
            } else {
                console.error(status, xhr.response)
            }
        };
        xhr.send();
    }

    function mkBtn(name) {
        let btn = document.createElement("button")
        btn.setAttribute('id', name)
        btn.innerText = name
        btn.setAttribute("data-state", "unsub")
        btn.classList.add('btnUnsub')

        btn.onclick = ev => {
            const that = ev.target
            const state = that.getAttribute("data-state")
            if ("unsub" === state) {
                ws.sub(name)
                that.setAttribute("data-state", "sub")
                that.classList.remove('btnUnsub')
                that.classList.add('btnSub')
            } else if ("sub" === state) {
                ws.unsub(name)
                that.setAttribute("data-state", "unsub")
                that.classList.remove('btnSub')
                that.classList.add('btnUnsub')
            } else if ("list" === state) {
                ws.tiles()
            }
        }
        return btn
    }
    function mkRqPlanesBtn(tileName, label) {
        let btn = document.createElement("button")
        btn.setAttribute('id', 'rq' + tileName + label)
        btn.innerText = label

        btn.onclick = () => {
            ws.currentPlanes(tileName)
        }
        return btn
    }

    getJSON("/grid", function (data) {
        const divLow = document.getElementById('grid-btns-low')
        const divHigh = document.getElementById('grid-btns-high')
        const divRqPlanes = document.getElementById('grid-btns-rq-planes')
        Object.keys(data).forEach(key => {
            let btnLow = mkBtn(key + "_low")
            let btnHigh = mkBtn(key + "_high")
            let btnRqPlanes = mkRqPlanesBtn(key, key)

            divLow.append(btnLow)
            divHigh.append(btnHigh)
            divRqPlanes.append(btnRqPlanes)
        })

        const divAll = document.getElementById("grid-btns-all")
        divAll.append(mkBtn("all_low"))
        divAll.append(mkBtn("all_high"))
        let list = mkBtn("list")
        list.setAttribute("data-state", "list")
        divAll.append(list)

        const unknownTileLoc = mkRqPlanesBtn("", "RQ All No Tile")
        divAll.append(unknownTileLoc)
    })

    function handleForm(id) {
        let f = document.getElementById(id)
        f.addEventListener("submit", function (e) {
            e.preventDefault()
            let elements = document.getElementById(id)

            let icao = ""
            let callSign = ""
            for (let i=0; i<elements.length; i++) {
                if ("icao" === elements[i].name) {
                    icao = elements[i].value;
                }
                if ("callsign" === elements[i].name) {
                    callSign = elements[i].value;
                }
            }

            ws.planeHistory(icao, callSign)
        })
    }
    handleForm('loc-hist')

})()