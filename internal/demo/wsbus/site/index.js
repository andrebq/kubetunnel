let open = false;
let websocket = new WebSocket("ws://" + location.host + "/bus", "protocolOne");
websocket.onopen = (ev) => { open = true; console.info("Websocket opened", ev); }
websocket.onerror = (ev) => { console.info("Websocket err", ev); }
websocket.onclose = (ev) => { open = false; console.info("Websocket closed", ev); }
websocket.onmessage = (ev) => { console.info(JSON.stringify(ev.data)); }


function tick() {
	if (open) {
		websocket.send(JSON.stringify({now: new Date().toISOString()}));
	}
	setTimeout(tick, 1000);
}
tick();
