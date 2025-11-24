import { useState, useEffect } from 'react';

export default function WebSocketTest() {
  const [connectionStatus, setConnectionStatus] = useState('Not connected');
  const [logs, setLogs] = useState([]);
  const [socket, setSocket] = useState(null);

  const addLog = (message) => {
    const timestamp = new Date().toLocaleTimeString();
    setLogs(prev => [...prev, `[${timestamp}] ${message}`]);
    console.log(`[WebSocket Test] ${message}`);
  };

  const testConnection = (userId = 1) => {
    addLog(`Attempting to connect with user ID: ${userId}`);
    setConnectionStatus('Connecting...');

    // Clean up existing connection
    if (socket) {
      socket.close();
    }

    const url = `ws://localhost:8080/ws?user_id=${userId}`;
    addLog(`WebSocket URL: ${url}`);
    
    const newSocket = new WebSocket(url);

    newSocket.onopen = (event) => {
      addLog('WebSocket connected successfully!');
      setConnectionStatus('Connected');
      setSocket(newSocket);
    };

    newSocket.onclose = (event) => {
      addLog(`WebSocket closed. Code: ${event.code}, Reason: ${event.reason || 'No reason'}`);
      setConnectionStatus(`Disconnected (${event.code})`);
    };

    newSocket.onerror = (event) => {
      addLog('WebSocket error occurred');
      setConnectionStatus('Error');
    };

    newSocket.onmessage = (event) => {
      addLog(`Message received: ${event.data}`);
    };
  };

  const sendTestMessage = () => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      const testMessage = {
        type: 'group_message',
        group_id: 1,
        from: 1,
        content: 'Test message',
        timestamp: new Date().toISOString()
      };
      socket.send(JSON.stringify(testMessage));
      addLog(`Sent test message: ${JSON.stringify(testMessage)}`);
    } else {
      addLog('Cannot send message - WebSocket not connected');
    }
  };

  const clearLogs = () => {
    setLogs([]);
  };

  return (
    <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
      <h1>WebSocket Connection Test</h1>
      
      <div style={{ marginBottom: '20px' }}>
        <p><strong>Status:</strong> {connectionStatus}</p>
        <button onClick={() => testConnection(1)} style={{ marginRight: '10px' }}>
          Test Connection (User ID: 1)
        </button>
        <button onClick={() => testConnection(2)} style={{ marginRight: '10px' }}>
          Test Connection (User ID: 2)
        </button>
        <button onClick={sendTestMessage} style={{ marginRight: '10px' }}>
          Send Test Message
        </button>
        <button onClick={clearLogs}>
          Clear Logs
        </button>
      </div>

      <div style={{ 
        border: '1px solid #ccc', 
        padding: '10px', 
        height: '400px', 
        overflow: 'auto', 
        backgroundColor: '#f9f9f9',
        fontFamily: 'monospace',
        fontSize: '12px'
      }}>
        {logs.map((log, index) => (
          <div key={index}>{log}</div>
        ))}
        {logs.length === 0 && <div style={{ color: '#999' }}>No logs yet...</div>}
      </div>
    </div>
  );
}