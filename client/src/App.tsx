import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [socket, setSocket] = useState<WebSocket | null>(null)
  const [connectionStatus, setConnectionStatus] = useState('未接続')
  const [message, setMessage] = useState('')
  const [messages, setMessages] = useState<string[]>([])

  useEffect(() => {
    const ws = new WebSocket(`ws://${window.location.hostname}:8080/ws`)

    ws.onopen = () => {
      console.log('WebSocket接続成功')
      setConnectionStatus('接続済み')
    }

    ws.onmessage = (event) => {
      console.log('メッセージ受信:', event.data)
      setMessages(prev => [...prev, `受信: ${event.data}`])
    }

    ws.onerror = (error) => {
      console.error('WebSocketエラー:', error)
      setConnectionStatus('エラー')
    }

    ws.onclose = () => {
      console.log('WebSocket切断')
      setConnectionStatus('切断')
    }

    setSocket(ws)

    return () => {
      ws.close()
    }
  }, [])

  const sendMessage = () => {
    if (socket && socket.readyState === WebSocket.OPEN && message) {
      socket.send(message)
      setMessages(prev => [...prev, `送信: ${message}`])
      setMessage('')
    }
  }

  return (
    <div className="App">
      <h1>Chess MMO - WebSocket テスト</h1>
      
      <div>
        <p>接続状態: <strong>{connectionStatus}</strong></p>
      </div>

      <div style={{ marginTop: '20px' }}>
        <input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
          placeholder="メッセージを入力"
          style={{ marginRight: '10px', padding: '5px' }}
        />
        <button onClick={sendMessage} style={{ padding: '5px 10px' }}>
          送信
        </button>
      </div>

      <div style={{ marginTop: '20px', textAlign: 'left', maxWidth: '600px', margin: '20px auto' }}>
        <h3>メッセージ履歴:</h3>
        <div style={{ 
          border: '1px solid #ccc', 
          padding: '10px', 
          height: '200px', 
          overflowY: 'auto',
          fontFamily: 'monospace',
          backgroundColor: '#f5f5f5',
          color: '#333'
        }}>
          {messages.map((msg, index) => (
            <div key={index} style={{ color: '#000' }}>{msg}</div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default App
