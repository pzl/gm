const SockProtocol = Object.freeze({
	Text: "text",
	Binary: "bin",
})

class Socket {
	constructor(url, protocol = SockProtocol.Text) {
		this.url = url
		this.protocol = protocol
		this.ws = null
		this.state = 'closed'
		this.onmessage = () => {} // set this to your own function to receive data
	}

	Connect() {
		this.Close()
		this.state = 'connecting'
		this.ws = new WebSocket(this.url.replace('http://', 'ws://'), this.protocol)
		if (this.protocol == 'bin') {
			this.ws.binaryType = 'arraybuffer'
		}
		this.ws.onopen = this.onopen.bind(this)
		this.ws.onmessage = this._onmessage.bind(this)
		this.ws.onerror = this.onerror.bind(this)
		this.ws.onclose = this.onclose.bind(this)
	}

	Close() {
		if (this.ws !== null) {
			this.ws.close()
		}
	}

	onopen(e) {
		console.log('websocket opened')
		this.state = 'connected'
	}
	onerror(e) {
		// ???
		console.log('ws error: ',e)
	}
	_onmessage(e) {
		this.onmessage(e)
	}
	onclose(e) {
		console.log('websocket closed')
		this.state = 'closed'
		this.ws = null
	}
}



export {
	Socket,
	SockProtocol,
}
