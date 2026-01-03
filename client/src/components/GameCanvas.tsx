import { useRef, useState, memo } from "react";
import type { GameState, GameConfig } from "../types";
import { GAME_CONSTANTS } from "../constants/game";
import { useMouseInput } from "../hooks/useMouseInput";
import { useTouchInput } from "../hooks/useTouchInput";
import { useKeyboardShortcuts } from "../hooks/useKeyboardShortcuts";
import { useGameRenderer } from "../hooks/useGameRenderer";

interface GameCanvasProps {
  gameState: GameState;
  playerId: string;
  gameConfig: GameConfig;
  onMouseMove: (x: number, y: number) => void;
  onMouseClick?: (x: number, y: number) => void;
}

const GameCanvas: React.FC<GameCanvasProps> = memo(
  ({ gameState, playerId, gameConfig, onMouseMove, onMouseClick }) => {
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const [showGrid, setShowGrid] = useState(true);
    const [showCulling, setShowCulling] = useState(false);

    // カスタムフックを使用
    useMouseInput({
      canvasRef,
      gameState,
      playerId,
      gameConfig,
      onMouseMove,
      onMouseClick,
    });

    // タッチ入力対応
    useTouchInput({
      canvasRef,
      gameState,
      playerId,
      gameConfig,
      onTouchMove: onMouseMove, // 同じ関数を使用
    });

    useKeyboardShortcuts({
      onToggleGrid: () => setShowGrid((prev) => !prev),
      onToggleCulling: () => setShowCulling((prev) => !prev),
    });

    useGameRenderer({
      canvasRef,
      gameState,
      playerId,
      gameConfig,
      showGrid,
      showCulling,
    });

    const { BASE_WIDTH, BASE_HEIGHT } = GAME_CONSTANTS;

    return (
      <canvas
        ref={canvasRef}
        width={BASE_WIDTH}
        height={BASE_HEIGHT}
        style={{
          display: "block",
          width: "100%",
          height: "100%",
          touchAction: "none", // タッチジェスチャーを無効化
        }}
      />
    );
  }
);

GameCanvas.displayName = "GameCanvas";

export default GameCanvas;
