import { useState, useEffect } from 'react'
import './App.css'

interface Message {
  id: string
  clientId: string
  text: string
  isSelf: boolean
}

function App() {
  const [socket, setSocket] = useState<WebSocket | null>(null)
  const [connectionStatus, setConnectionStatus] = useState('未接続')
  const [message, setMessage] = useState('')
  const [messages, setMessages] = useState<Message[]>([])
  const [clientId] = useState(`${Date.now()}-${Math.random().toString(36).substring(2, 9)}`)

  useEffect(() => {
    const ws = new WebSocket(`ws://${window.location.hostname}:8080/ws`)

    ws.onopen = () => {
      console.log('WebSocket接続成功')
      setConnectionStatus('接続済み')
    }

    ws.onmessage = (event) => {
      console.log('メッセージ受信:', event.data)
      try {
        const data = JSON.parse(event.data)
        setMessages(prev => [...prev, {
          id: Date.now().toString() + Math.random(),
          clientId: data.clientId,
          text: data.text,
          isSelf: data.clientId === clientId
        }])
      } catch (e) {
        // 古い形式のメッセージ対応
        setMessages(prev => [...prev, {
          id: Date.now().toString() + Math.random(),
          clientId: '不明',
          text: event.data,
          isSelf: false
        }])
      }
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
      const msgData = {
        clientId: clientId,
        text: message
      }
      socket.send(JSON.stringify(msgData))
      setMessage('')
    }
  }

  return (
    <div className="App">
      <h1>チャット</h1>
      
      <div>
        <p>接続: <strong>{connectionStatus}</strong></p>
      </div>

      <div style={{ marginTop: '20px', textAlign: 'left', maxWidth: '600px', margin: '20px auto' }}>
        <div style={{ 
          border: '1px solid #ccc', 
          padding: '10px', 
          height: '400px', 
          overflowY: 'auto',
          backgroundColor: '#f5f5f5',
          display: 'flex',
          flexDirection: 'column'
        }}>
          {messages.map((msg) => (
            <div 
              key={msg.id} 
              style={{ 
                marginBottom: '10px',
                textAlign: msg.isSelf ? 'right' : 'left'
              }}
            >
              <div style={{ 
                display: 'inline-block',
                padding: '8px 12px',
                borderRadius: '10px',
                backgroundColor: msg.isSelf ? '#007bff' : '#e9ecef',
                color: msg.isSelf ? 'white' : '#000',
                maxWidth: '70%'
              }}>
                <div>{msg.text}</div>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div style={{ marginTop: '20px', maxWidth: '600px', margin: '20px auto' }}>
        <input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
          placeholder="メッセージを入力..."
          style={{ 
            width: '80%',
            padding: '10px',
            fontSize: '16px',
            borderRadius: '5px',
            border: '1px solid #ccc'
          }}
        />
        <button 
          onClick={sendMessage} 
          style={{ 
            marginLeft: '10px',
            padding: '10px 20px',
            fontSize: '16px',
            backgroundColor: '#007bff',
            color: 'white',
            border: 'none',
            borderRadius: '5px',
            cursor: 'pointer'
          }}
        >
          送信
        </button>
      </div>
    </div>
  )
}

export default App
