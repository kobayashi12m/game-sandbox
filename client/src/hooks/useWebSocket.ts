import { useEffect, useRef, useState, useCallback } from 'react';
import type { 
  GameState, 
  JoinMessage, 
  AccelerationMessage,
  WebSocketMessage,
  GameConfig,
  GameConfigMessage,
  ScoreInfo,
  ScoreboardMessage
} from '../types';
import { DEFAULT_GAME_CONFIG } from '../types';

interface UseWebSocketProps {
  roomId: string;
  playerName: string;
  isConnected: boolean;
}

interface UseWebSocketReturn {
  gameState: GameState;
  playerId: string;
  sendAcceleration: (x: number, y: number) => void;
  sendEjectSatellite: (targetX: number, targetY: number) => void;
  isConnecting: boolean;
  gameConfig: GameConfig;
  scoreboard: ScoreInfo[];
  myScore: ScoreInfo | null;
}

export const useWebSocket = ({ 
  roomId, 
  playerName, 
  isConnected 
}: UseWebSocketProps): UseWebSocketReturn => {
  const [gameState, setGameState] = useState<GameState>({ pls: [] });
  const [playerId, setPlayerId] = useState<string>('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [gameConfig, setGameConfig] = useState<GameConfig>(DEFAULT_GAME_CONFIG);
  const [scoreboard, setScoreboard] = useState<ScoreInfo[]>([]);
  const [myScore, setMyScore] = useState<ScoreInfo | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  // メッセージハンドラー
  const handleMessage = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case 'gameJoined':
        if ('playerId' in message && typeof message.playerId === 'string') {
          setPlayerId(message.playerId);
        }
        break;
      
      case 'gameState':
        if ('state' in message) {
          setGameState(message.state as GameState);
        }
        break;
        
      case 'gameConfig':
        if ('config' in message) {
          const configMsg = message as GameConfigMessage;
          setGameConfig(configMsg.config);
        }
        break;

      case 'scoreboard':
        if ('scoreboard' in message) {
          const scoreboardMsg = message as ScoreboardMessage;
          setScoreboard(scoreboardMsg.scoreboard);
          
          // 自分のスコア情報を保存
          if (scoreboardMsg.myScore) {
            setMyScore(scoreboardMsg.myScore);
          }
        }
        break;
    }
  }, []);

  // WebSocket接続の確立
  useEffect(() => {
    if (!isConnected) return;

    setIsConnecting(true);
    
    const wsUrl = window.location.hostname === 'localhost' 
      ? 'ws://localhost:8081/ws' 
      : `ws://${window.location.hostname}:8081/ws`;
    
    const websocket = new WebSocket(wsUrl);
    wsRef.current = websocket;

    websocket.onopen = () => {
      setIsConnecting(false);
      
      const joinMessage: JoinMessage = {
        type: 'join',
        roomId,
        playerName
      };
      websocket.send(JSON.stringify(joinMessage));
    };

    websocket.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        handleMessage(message);
      } catch (error) {
        console.error('📡 JSON parse error:', error, event.data);
      }
    };

    websocket.onclose = () => {
      setIsConnecting(false);
      setPlayerId('');
      setGameState({ pls: [] });
      setScoreboard([]);
    };

    websocket.onerror = () => {
      setIsConnecting(false);
    };

    return () => {
      websocket.close();
      wsRef.current = null;
    };
  }, [isConnected, roomId, playerName, handleMessage]);


  // 加速度送信（360度自由移動用）
  const sendAcceleration = useCallback((x: number, y: number) => {
    const websocket = wsRef.current;
    if (!websocket || websocket.readyState !== WebSocket.OPEN) return;

    const accelerationMessage: AccelerationMessage = {
      type: 'setAcceleration',
      x,
      y
    };
    
    websocket.send(JSON.stringify(accelerationMessage));
  }, []);



  // 衛星射出メッセージの送信
  const sendEjectSatellite = useCallback((targetX: number, targetY: number) => {
    const websocket = wsRef.current;
    if (!websocket || websocket.readyState !== WebSocket.OPEN) return;

    const ejectMessage = {
      type: 'ejectSatellite',
      targetX,
      targetY
    };
    
    websocket.send(JSON.stringify(ejectMessage));
  }, []);

  return {
    gameState,
    playerId,
    sendAcceleration,
    sendEjectSatellite,
    isConnecting,
    gameConfig,
    scoreboard,
    myScore
  };
};