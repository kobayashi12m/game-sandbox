import { useEffect } from 'react';
import type { GameState, GameConfig } from '../types';
import { GAME_CONSTANTS } from '../constants/game';
import { UI_CONFIG } from '../constants/ui';
import { drawGame } from '../components/drawing/GameRenderer';

interface UseGameRendererProps {
  canvasRef: React.RefObject<HTMLCanvasElement | null>;
  gameState: GameState;
  playerId: string;
  gameConfig: GameConfig;
  showGrid: boolean;
  showCulling: boolean;
}

export const useGameRenderer = ({
  canvasRef,
  gameState,
  playerId,
  gameConfig,
  showGrid,
  showCulling
}: UseGameRendererProps) => {
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const { BASE_WIDTH, BASE_HEIGHT } = GAME_CONSTANTS;
    const canvasSize = { width: BASE_WIDTH, height: BASE_HEIGHT };

    try {
      if (!gameState || !gameState.pls) return;

      drawGame(
        ctx,
        gameState,
        playerId,
        gameConfig,
        canvasSize,
        showGrid,
        showCulling,
        UI_CONFIG.SHOW_LEFT_UI
      );
    } catch (error) {
      console.error("🚨 DRAW ERROR:", error);
      console.error("GameState:", gameState);
      console.error("PlayerID:", playerId);
    }
  }, [gameState, playerId, gameConfig, showGrid, showCulling]);
};