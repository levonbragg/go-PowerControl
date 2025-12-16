// Go PowerControl - Vanilla JavaScript Application
const app = {
    devices: [],
    messages: [],
    selectedDevice: null,
    connected: false,

    async init() {
        console.log('Initializing Go PowerControl...');

        // Load initial data
        await this.loadDevices();
        await this.loadMessages();
        this.connected = await window.go.app.App.GetConnectionStatus();
        this.updateConnectionStatus(this.connected);

        // Check if config is empty
        const isEmpty = await window.go.app.App.IsConfigEmpty();
        if (isEmpty) {
            this.showSetup();
        }

        // Subscribe to events
        window.runtime.EventsOn('device:update', () => {
            this.loadDevices();
        });

        window.runtime.EventsOn('message:new', () => {
            this.loadMessages();
        });

        window.runtime.EventsOn('connection:status', (isConnected) => {
            this.connected = isConnected;
            this.updateConnectionStatus(isConnected);
        });

        window.runtime.EventsOn('log:cleared', () => {
            this.messages = [];
            this.renderMessages();
        });

        console.log('App initialized');
    },

    async loadDevices() {
        try {
            this.devices = await window.go.app.App.GetDevices();
            this.renderDevices();
        } catch (error) {
            console.error('Failed to load devices:', error);
        }
    },

    async loadMessages() {
        try {
            this.messages = await window.go.app.App.GetMessages();
            this.renderMessages();
        } catch (error) {
            console.error('Failed to load messages:', error);
        }
    },

    renderDevices() {
        const tbody = document.getElementById('deviceTableBody');

        if (this.devices.length === 0) {
            tbody.innerHTML = '<tr><td colspan="3" style="text-align: center; color: var(--text-secondary); padding: 2rem;">No devices found</td></tr>';
            return;
        }

        let html = '';
        let lastDevice = '';

        this.devices.forEach((device, index) => {
            const showDevice = device.deviceName !== lastDevice;
            lastDevice = device.deviceName;

            const statusClass = device.status === 'ON' ? 'status-on' : 'status-off';

            html += `<tr onclick="app.selectDevice(${index})">
                <td>${showDevice ? device.deviceName : ''}</td>
                <td>${device.outletNumber}</td>
                <td class="${statusClass}">${device.status}</td>
            </tr>`;
        });

        tbody.innerHTML = html;
    },

    renderMessages() {
        const messageList = document.getElementById('messageList');

        if (this.messages.length === 0) {
            messageList.innerHTML = '<div style="color: var(--text-secondary); text-align: center; padding: 2rem;">No messages yet</div>';
            return;
        }

        let html = '';
        this.messages.forEach(msg => {
            const time = new Date(msg.timestamp).toLocaleTimeString();
            const direction = msg.direction === 'Send' ? '>>' : '<<';
            const className = msg.direction === 'Send' ? 'message-send' : 'message-recv';

            html += `<div class="message-item ${className}">[${time}] ${direction} ${msg.direction}: ${msg.topic} ${msg.payload}</div>`;
        });

        messageList.innerHTML = html;
        messageList.scrollTop = 0;
    },

    selectDevice(index) {
        this.selectedDevice = this.devices[index];

        document.getElementById('selectedDevice').textContent = this.selectedDevice.deviceName;
        document.getElementById('selectedOutlet').textContent = this.selectedDevice.outletNumber;
        document.getElementById('stateSelector').value = this.selectedDevice.status;
        document.getElementById('stateSelector').disabled = false;
        document.getElementById('sendButton').disabled = false;

        const rows = document.querySelectorAll('#deviceTableBody tr');
        rows.forEach((row, i) => {
            if (i === index) {
                row.classList.add('selected');
            } else {
                row.classList.remove('selected');
            }
        });
    },

    async handleSearch(searchText) {
        try {
            if (searchText) {
                this.devices = await window.go.app.App.SearchDevices(searchText);
            } else {
                this.devices = await window.go.app.App.GetDevices();
            }
            this.renderDevices();
        } catch (error) {
            console.error('Search failed:', error);
        }
    },

    async sendCommand() {
        if (!this.selectedDevice) return;

        const state = document.getElementById('stateSelector').value;

        try {
            await window.go.app.App.SendCommand(
                this.selectedDevice.deviceName,
                this.selectedDevice.outletNumber,
                state
            );
        } catch (error) {
            alert('Failed to send command: ' + error);
        }
    },

    async disconnect() {
        try {
            await window.go.app.App.Disconnect();
        } catch (error) {
            console.error('Disconnect failed:', error);
        }
    },

    async clearLog() {
        try {
            await window.go.app.App.ClearLog();
        } catch (error) {
            console.error('Clear log failed:', error);
        }
    },

    exit() {
        window.runtime.Quit();
    },

    async showSetup() {
        try {
            const config = await window.go.app.App.GetConfig();

            document.getElementById('setupUsername').value = config.username || '';
            document.getElementById('setupServer').value = config.mqttServer || '';
            document.getElementById('setupPort').value = config.serverPort || 1883;
            document.getElementById('setupSubscribe').value = config.subscribeString || 'power/#';
            document.getElementById('setupPassword').value = '';

            document.getElementById('setupDialog').style.display = 'flex';
        } catch (error) {
            console.error('Failed to load config:', error);
        }
    },

    closeSetup() {
        document.getElementById('setupDialog').style.display = 'none';
    },

    async saveSettings() {
        const username = document.getElementById('setupUsername').value;
        const password = document.getElementById('setupPassword').value;
        const server = document.getElementById('setupServer').value;
        const port = parseInt(document.getElementById('setupPort').value);
        const subscribeString = document.getElementById('setupSubscribe').value;

        try {
            await window.go.app.App.SaveSettings(username, password, server, port, subscribeString);
            this.closeSetup();
            this.connected = await window.go.app.App.GetConnectionStatus();
            this.updateConnectionStatus(this.connected);
            await this.loadDevices();
        } catch (error) {
            alert('Failed to save settings: ' + error);
        }
    },

    showAbout() {
        document.getElementById('aboutDialog').style.display = 'flex';
    },

    closeAbout() {
        document.getElementById('aboutDialog').style.display = 'none';
    },

    updateConnectionStatus(connected) {
        const indicator = document.getElementById('statusIndicator');
        const text = document.getElementById('statusText');

        if (connected) {
            indicator.classList.add('connected');
            text.textContent = 'Connected';
        } else {
            indicator.classList.remove('connected');
            text.textContent = 'Disconnected';
        }
    }
};

document.addEventListener('DOMContentLoaded', () => {
    app.init();
});
