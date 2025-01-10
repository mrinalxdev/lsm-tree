let ws;
let wsReconnectTimer;

function connectWebSocket() {
    clearTimeout(wsReconnectTimer);

    ws = new WebSocket(`ws://${window.location.host}/ws`);

    ws.onopen = () => {
        console.log('WebSocket connected');
        updateStatus('Connected to server');
        enableControls(true);
    };

    ws.onclose = () => {
        console.log('WebSocket disconnected');
        updateStatus('Disconnected from server - Reconnecting...', true);
        enableControls(false);
        wsReconnectTimer = setTimeout(connectWebSocket, 3000);
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        updateStatus('Connection error', true);
        enableControls(false);
    };

    ws.onmessage = (event) => {
        try {
            // First try to parse the message
            const message = JSON.parse(event.data);

            // Handle the visualization update
            if (typeof handleWSMessage === 'function') {
                handleWSMessage(message);
            }

            // Update status based on message type
            switch (message.type) {
                case 'set':
                    updateStatus(`Set ${message.key} = ${message.value}`);
                    break;
                case 'get':
                    updateStatus(`Got ${message.key} = ${message.value}`);
                    break;
                case 'delete':
                    updateStatus(`Deleted ${message.key}`);
                    break;
                case 'error':
                    updateStatus(message.value, true);
                    break;
            }

            // Add to operation log with proper stringification
            const logEntry = `${message.type.toUpperCase()}: ${message.key}${message.value ? ` = ${message.value}` : ''}`;
            addToLog(logEntry);
        } catch (error) {
            console.error('Error handling message:', error, 'Raw message:', event.data);
            updateStatus('Error processing server response', true);
        }
    };
}

function enableControls(enabled) {
    const buttons = document.querySelectorAll('.operation-button');
    buttons.forEach(button => {
        button.disabled = !enabled;
    });
}

function sendMessage(type, key, value = null) {
    if (!key) {
        updateStatus('Key is required', true);
        return;
    }

    if (type === 'set' && !value) {
        updateStatus('Value is required for SET operation', true);
        return;
    }

    if (ws && ws.readyState === WebSocket.OPEN) {
        const message = {
            type: type,
            key: key,
            value: value,
            timestamp: new Date().toISOString()
        };
        ws.send(JSON.stringify(message));
    } else {
        updateStatus('Not connected to server', true);
    }
}

function handleSet() {
    const key = document.getElementById('keyInput').value.trim();
    const value = document.getElementById('valueInput').value.trim();

    if (!key || !value) {
        updateStatus('Both key and value are required', true);
        return;
    }

    sendMessage('set', key, value);
    document.getElementById('keyInput').value = '';
    document.getElementById('valueInput').value = '';
}

function handleGet() {
    const key = document.getElementById('keyInput').value.trim();

    if (!key) {
        updateStatus('Key is required', true);
        return;
    }

    sendMessage('get', key);
}

function handleDelete() {
    const key = document.getElementById('keyInput').value.trim();

    if (!key) {
        updateStatus('Key is required', true);
        return;
    }

    sendMessage('delete', key);
    document.getElementById('keyInput').value = '';
}

function addToLog(message) {
    const logContainer = document.getElementById('logContainer');
    const logEntry = document.createElement('div');
    logEntry.className = 'py-1 border-b border-gray-200';
    logEntry.textContent = `${new Date().toLocaleTimeString()} - ${message}`;
    logContainer.appendChild(logEntry);
    logContainer.scrollTop = logContainer.scrollHeight;
}
