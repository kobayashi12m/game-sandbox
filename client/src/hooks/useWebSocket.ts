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

// 接続タイムアウト
const CONNECTION_TIMEOUT_MS = 3000;

export const useWebSocket = ({
  roomId,
  playerName,
  isConnected,
}: UseWebSocketProps): UseWebSocketReturn => {
  const [gameState, setGameState] = useState<GameState>({ pls: [] });
  const [playerId, setPlayerId] = useState<string>("");
  const [isConnecting, setIsConnecting] = useState(false);
  const [gameConfig, setGameConfig] = useState<GameConfig>(DEFAULT_GAME_CONFIG);
  const [scoreboard, setScoreboard] = useState<ScoreInfo[]>([]);
  const [myScore, setMyScore] = useState<ScoreInfo | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectingRef = useRef(false);

  // 接続先URL
  const wsUrl =
    import.meta.env.VITE_WS_URL ||
    `ws://${window.location.hostname}:8081/ws`;

  // メッセージハンドラー
  const handleMessage = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case "gameJoined":
        if ("playerId" in message && typeof message.playerId === "string") {
          setPlayerId(message.playerId);
        }
        break;

      case "gameState":
        if ("state" in message) {
          setGameState(message.state as GameState);
        }
        break;

      case "gameConfig":
        if ("config" in message) {
          const configMsg = message as GameConfigMessage;
          setGameConfig(configMsg.config);
        }
        break;

      case "scoreboard":
        if ("scoreboard" in message) {
          const scoreboardMsg = message as ScoreboardMessage;
          setScoreboard(scoreboardMsg.scoreboard);
          if (scoreboardMsg.myScore) {
            setMyScore(scoreboardMsg.myScore);
          }
        }
        break;
    }
  }, []);

  // WebSocket接続を作成
  const createConnection = useCallback(() => {
    // 既存の接続があれば閉じる
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    console.log("[WS] Creating connection:", wsUrl);
    setIsConnecting(true);

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    // タイムアウト処理（iOS Safari対策）
    const timeoutId = setTimeout(() => {
      if (ws.readyState === WebSocket.CONNECTING) {
        console.log("[WS] Connection timeout, retrying...");
        ws.close();
        if (!reconnectingRef.current) {
          reconnectingRef.current = true;
          createConnection();
          reconnectingRef.current = false;
        }
      }
    }, CONNECTION_TIMEOUT_MS);

    ws.onopen = () => {
      console.log("[WS] Connected");
      clearTimeout(timeoutId);
      setIsConnecting(false);

      const joinMessage: JoinMessage = {
        type: "join",
        roomId,
        playerName,
      };
      ws.send(JSON.stringify(joinMessage));
    };

    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        handleMessage(message);
      } catch (error) {
        console.error("[WS] Parse error:", error);
      }
    };

    ws.onclose = (event) => {
      console.log("[WS] Closed:", event.code, event.reason);
      clearTimeout(timeoutId);
      setIsConnecting(false);
      setPlayerId("");
      setGameState({ pls: [] });
      setScoreboard([]);
    };

    ws.onerror = () => {
      console.log("[WS] Error");
      clearTimeout(timeoutId);
      setIsConnecting(false);
    };

    return () => {
      clearTimeout(timeoutId);
    };
  }, [wsUrl, roomId, playerName, handleMessage]);

  // メイン接続エフェクト
  useEffect(() => {
    if (!isConnected) {
      return;
    }

    const cleanup = createConnection();

    // ページ離脱時にWebSocketを閉じる
    const handlePageHide = () => {
      console.log("[WS] Page hide, closing");
      wsRef.current?.close();
    };

    // ページ復帰時に再接続
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        const ws = wsRef.current;
        if (!ws || ws.readyState === WebSocket.CLOSED) {
          console.log("[WS] Page visible, reconnecting...");
          createConnection();
        }
      }
    };

    window.addEventListener("pagehide", handlePageHide);
    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      cleanup?.();
      window.removeEventListener("pagehide", handlePageHide);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [isConnected, createConnection]);

  // 加速度送信
  const sendAcceleration = useCallback((x: number, y: number) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    const message: AccelerationMessage = {
      type: "setAcceleration",
      x,
      y,
    };
    ws.send(JSON.stringify(message));
  }, []);

  // 衛星射出
  const sendEjectSatellite = useCallback((targetX: number, targetY: number) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    const message = {
      type: "ejectSatellite",
      targetX,
      targetY,
    };
    ws.send(JSON.stringify(message));
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
