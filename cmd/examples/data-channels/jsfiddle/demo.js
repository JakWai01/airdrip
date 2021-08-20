/* eslint-env browser */

let pc = new RTCPeerConnection({
    iceServer: [
	{
	    urls: 'stun:stun.l.google.com:19302'
	}
    ]
})

// Add message to logs
let log = msg => {
    document.getElementById('logs').innerHTML += msg + '<br>'
}

// Create a new channel
let sendChannel = pc.createDataChannel('foo')

// Log message on close
sendChannel.onclose = () => console.log('sendChannel has closed')

// Log message on open
sendChannel.onopen = () => console.log('sendChanel has opened')

// Log message on message
sendChannel.onmessage = e => log(`Message from DataChannel '${sendChannel.label}' payload '${e.data}'`)

// Called when ICE connection state changes
pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)

// Called on ICE candidate
pc.onicecandidate = event => {
    if (event.candidate === null) {
        // btoa encodes string to base-64
	    document.getElementById('localSessionDescription').value = btoa(JSON.stringify(pc.localDescription))
    }
}

// Called when negotiation is needed
pc.onnegotiationneeded = e => pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)

// Send message 
window.sendMessage = () => {
    let message = document.getElementById('message').value
    if (message === '') {
	    return alert('Message must not be empty')
    }

    sendChannel.send(message)
}

// Start connection
window.startSession = () => {
    let sd = document.getElementById('remoteSessionDescription').value
    if (sd === '') {
	    return alert('Session Description must not be empty')
    }

    try {
        // encode to base-64 to string
	    pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sd))))
    } catch(e) {
	    alert(e)
    }
}
