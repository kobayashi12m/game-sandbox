import { useState, useEffect, useRef } from 'react'
import './App.css'
import GameCanvas from './GameCanvas'

interface Player {
  id: string;
  x: number;
  y: number;
  coreSize: number;
  guardSize: number;
}

interface GameMessage {
  type: 'chat' | 'gameInit' | 'gameState' | 'move' | 'playerDisconnect';
  data: any;
}

function App() {
  const [socket, setSocket] = useState<WebSocket | null>(null)
  const [connectionStatus, setConnectionStatus] = useState('未接続')
  const [player, setPlayer] = useState<Player | null>(null)
  const [otherPlayers, setOtherPlayers] = useState<Map<string, Player>>(new Map())
  const lastUpdateTime = useRef(0)
  const playerIdRef = useRef<string | null>(null)

  useEffect(() => {
    const ws = new WebSocket(`ws://${window.location.hostname}:8080/ws`)

    ws.onopen = () => {
      console.log('WebSocket接続成功')
      setConnectionStatus('接続済み')
    }

    ws.onmessage = (event) => {
      try {
        const gameMsg: GameMessage = JSON.parse(event.data)
        
        if (gameMsg.type === 'gameInit') {
          // ゲーム初期化メッセージを受信
          const playerData: Player = gameMsg.data
          setPlayer(playerData)
          playerIdRef.current = playerData.id
        } else if (gameMsg.type === 'gameState') {
          // 他のプレイヤーの状態更新
          const playerData: Player = gameMsg.data
          
          if (playerIdRef.current && playerData.id !== playerIdRef.current) {
            setOtherPlayers(prev => {
              const newMap = new Map(prev)
              newMap.set(playerData.id, playerData)
              return newMap
            })
          }
        } else if (gameMsg.type === 'playerDisconnect') {
          // プレイヤー切断
          const playerId: string = gameMsg.data
          setOtherPlayers(prev => {
            const newMap = new Map(prev)
            newMap.delete(playerId)
            return newMap
          })
        }
      } catch (e) {
        // エラーは無視（パフォーマンス向上のため）
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

  // 位置更新ハンドラー（レート制限付き）
  const handlePositionUpdate = (x: number, y: number) => {
    const now = Date.now()
    // 100ms（10FPS）のレート制限に変更
    if (now - lastUpdateTime.current < 100) return
    
    if (socket && socket.readyState === WebSocket.OPEN && playerIdRef.current) {
      const moveMsg: GameMessage = {
        type: 'move',
        data: {
          id: playerIdRef.current,
          x: x,
          y: y
        }
      }
      socket.send(JSON.stringify(moveMsg))
      lastUpdateTime.current = now
    }
  }

  return (
    <div className="App">
      <div>
        <p>接続: <strong>{connectionStatus}</strong></p>
        {player && <p>プレイヤーID: <strong>{player.id}</strong></p>}
      </div>

      <GameCanvas 
        player={player} 
        otherPlayers={Array.from(otherPlayers.values())}
        onPositionUpdate={handlePositionUpdate} 
      />
    </div>
  )
}

export default App