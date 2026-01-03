import { useEffect, useRef } from 'react';
import type { GameState, GameConfig } from '../types';
import { getPlayer } from '../types';
import { GAME_CONSTANTS } from '../constants/game';
import { convertMouseToWorldCoords } from '../utils/mouse';

interface UseMouseInputProps {
  canvasRef: React.RefObject<HTMLCanvasElement | null>;
  gameState: GameState;
  playerId: string;
  gameConfig: GameConfig;
  onMouseMove: (x: number, y: number) => void;
  onMouseClick?: (x: number, y: number) => void;
}

export const useMouseInput = ({
  canvasRef,
  gameState,
  playerId,
  gameConfig,
  onMouseMove,
  onMouseClick
}: UseMouseInputProps) => {
  const lastSendTime = useRef(0);
  const pendingMouseEvent = useRef<MouseEvent | null>(null);
  const animationFrameId = useRef<number | null>(null);

  // マウスムーブの処理
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const handleMouseMove = (event: MouseEvent) => {
      if (!gameState) return;

      const THROTTLE_DELAY = 16; // 60fps
      const now = Date.now();
      
      pendingMouseEvent.current = event;

      if (now - lastSendTime.current >= THROTTLE_DELAY) {
        processMouseEvent();
      } else if (!animationFrameId.current) {
        animationFrameId.current = requestAnimationFrame(processMouseEvent);
      }
    };

    const processMouseEvent = () => {
      const event = pendingMouseEvent.current;
      if (!event || !canvas) return;

      const now = Date.now();
      if (now - lastSendTime.current < 16) return; // 最小間隔

      pendingMouseEvent.current = null;
      lastSendTime.current = now;
      const rect = canvas.getBoundingClientRect();
      const currentPlayer = gameState.pls?.find((p) => p[0] === playerId);
      const playerPosition = currentPlayer
        ? getPlayer(currentPlayer)?.cel?.c?.p
        : undefined;

      if (!playerPosition) return;

      // カメラズーム設定を取得
      const gameZoomScale = gameConfig.cameraZoomScale || 1.0;

      const worldCoords = convertMouseToWorldCoords(event, rect, playerPosition, gameZoomScale);
      if (!worldCoords) return;
      const { worldX, worldY } = worldCoords;

      // コアからマウス位置への方向ベクトルを計算
      const dx = worldX - playerPosition.x;
      const dy = worldY - playerPosition.y;
      const distance = Math.sqrt(dx * dx + dy * dy);

      // 距離に応じた速度制御
      const { MIN_DISTANCE: minDistance, MAX_DISTANCE: maxDistance } =
        GAME_CONSTANTS;

      if (distance > minDistance) {
        // 正規化された方向ベクトル
        const normalizedX = dx / distance;
        const normalizedY = dy / distance;

        // 距離に応じて強度を調整（線形補間）
        const clampedDistance = Math.min(distance, maxDistance);
        const intensity = (clampedDistance - minDistance) / (maxDistance - minDistance);

        const accelerationX = normalizedX * intensity;
        const accelerationY = normalizedY * intensity;

        onMouseMove(accelerationX, accelerationY);
      } else {
        onMouseMove(0, 0);
      }

      if (animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
        animationFrameId.current = null;
      }
    };

    canvas.addEventListener("mousemove", handleMouseMove);
    return () => {
      canvas.removeEventListener("mousemove", handleMouseMove);
      if (animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
      }
    };
  }, [gameState, playerId, gameConfig.cameraZoomScale, onMouseMove]);

  // マウスクリックの処理
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !onMouseClick) return;

    const handleClick = (event: MouseEvent) => {
      const rect = canvas.getBoundingClientRect();
      const currentPlayer = gameState.pls?.find((p) => p[0] === playerId);
      const playerPosition = currentPlayer
        ? getPlayer(currentPlayer)?.cel?.c?.p
        : undefined;

      if (!playerPosition) return;

      // カメラズーム設定を取得
      const gameZoomScale = gameConfig.cameraZoomScale || 1.0;

      const worldCoords = convertMouseToWorldCoords(event, rect, playerPosition, gameZoomScale);
      if (!worldCoords) return;
      const { worldX, worldY } = worldCoords;

      onMouseClick(worldX, worldY);
    };

    canvas.addEventListener("click", handleClick);
    return () => {
      canvas.removeEventListener("click", handleClick);
    };
  }, [gameState, playerId, gameConfig.cameraZoomScale, onMouseClick]);
};