import { useEffect, useRef, useState, useCallback } from "react";
import type {
  GameState,
  JoinMessage,
  AccelerationMessage,
  WebSocketMessage,
  GameConfig,
  GameConfigMessage,
  ScoreInfo,
  ScoreboardMessage,
} from "../types";
import { DEFAULT_GAME_CONFIG } from "../types";

interface UseWebSocketProps {
  roomId: string;
  playerName: string;
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
}: UseWebSocketProps): UseWebSocketReturn => {
  const [gameState, setGameState] = useState<GameState>({ pls: [] });
  const [playerId, setPlayerId] = useState<string>("");
  const [isConnecting, setIsConnecting] = useState(false);
  const [gameConfig, setGameConfig] = useState<GameConfig>(DEFAULT_GAME_CONFIG);
  const [scoreboard, setScoreboard] = useState<ScoreInfo[]>([]);
  const [myScore, setMyScore] = useState<ScoreInfo | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const wsUrl =
    import.meta.env.VITE_WS_URL || `${wsProtocol}//${window.location.host}/ws`;

  // 状態をリセット
  const resetState = useCallback(() => {
    setPlayerId("");
    setGameState({ pls: [] });
    setScoreboard([]);
    setIsConnecting(false);
  }, []);

  // 接続を完全にクリア
  const clearConnection = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.onopen = null;
      wsRef.current.onmessage = null;
      wsRef.current.onclose = null;
      wsRef.current.onerror = null;
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  // 新規接続
  const connect = useCallback(() => {
    clearConnection();
    resetState();
    setIsConnecting(true);

    console.log("[WS] Connecting:", wsUrl);
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("[WS] Open");
      ws.send(JSON.stringify({ type: "join", roomId, playerName } as JoinMessage));
    };

    ws.onmessage = (e) => {
      try {
        const msg: WebSocketMessage = JSON.parse(e.data);
        if (msg.type === "gameJoined" && "playerId" in msg) {
          console.log("[WS] Joined:", msg.playerId);
          setPlayerId(msg.playerId as string);
          setIsConnecting(false);
        } else if (msg.type === "gameState" && "state" in msg) {
          setGameState(msg.state as GameState);
        } else if (msg.type === "gameConfig" && "config" in msg) {
          setGameConfig((msg as GameConfigMessage).config);
        } else if (msg.type === "scoreboard" && "scoreboard" in msg) {
          const sm = msg as ScoreboardMessage;
          setScoreboard(sm.scoreboard);
          if (sm.myScore) setMyScore(sm.myScore);
        }
      } catch (err) {
        console.error("[WS] Parse error:", err);
      }
    };

    ws.onclose = (e) => {
      console.log("[WS] Close:", e.code);
      resetState();
      // 異常終了なら再接続
      if (e.code === 1006 || e.code === 1005) {
        setTimeout(connect, 500);
      }
    };

    ws.onerror = () => {
      console.log("[WS] Error");
    };
  }, [wsUrl, roomId, playerName, clearConnection, resetState]);

  // メインエフェクト
  useEffect(() => {
    connect();

    return () => {
      clearConnection();
    };
  }, [connect, clearConnection]);

  const sendAcceleration = useCallback((x: number, y: number) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({ type: "setAcceleration", x, y } as AccelerationMessage)
      );
    }
  }, []);

  const sendEjectSatellite = useCallback((targetX: number, targetY: number) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({ type: "ejectSatellite", targetX, targetY })
      );
    }
  }, []);

  return {
    gameState,
    playerId,
    sendAcceleration,
    sendEjectSatellite,
    isConnecting,
    gameConfig,
    scoreboard,
    myScore,
  };
};
