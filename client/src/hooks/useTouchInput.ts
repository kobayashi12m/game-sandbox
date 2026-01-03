import { useEffect, useRef } from 'react';
import type { GameState, GameConfig } from '../types';
import { getPlayer } from '../types';
import { GAME_CONSTANTS } from '../constants/game';
import { convertMouseToWorldCoords } from '../utils/mouse';

interface UseTouchInputProps {
  canvasRef: React.RefObject<HTMLCanvasElement | null>;
  gameState: GameState;
  playerId: string;
  gameConfig: GameConfig;
  onTouchMove: (x: number, y: number) => void;
  onTouchClick?: (x: number, y: number) => void;
}

export const useTouchInput = ({
  canvasRef,
  gameState,
  playerId,
  gameConfig,
  onTouchMove,
  onTouchClick
}: UseTouchInputProps) => {
  const lastSendTime = useRef(0);
  const pendingTouchEvent = useRef<TouchEvent | null>(null);
  const animationFrameId = useRef<number | null>(null);
  const isTouching = useRef(false);

  // タッチからマウス形式への変換
  const touchToMouseEvent = (touch: Touch): MouseEvent => {
    return {
      clientX: touch.clientX,
      clientY: touch.clientY,
    } as MouseEvent;
  };

  // タッチムーブの処理
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const handleTouchMove = (event: TouchEvent) => {
      event.preventDefault(); // スクロール防止
      if (!gameState || event.touches.length === 0 || !isTouching.current) return;

      const THROTTLE_DELAY = 16; // 60fps
      const now = Date.now();
      
      pendingTouchEvent.current = event;

      if (now - lastSendTime.current >= THROTTLE_DELAY) {
        processTouchEvent();
      } else if (!animationFrameId.current) {
        animationFrameId.current = requestAnimationFrame(processTouchEvent);
      }
    };

    const processTouchEvent = () => {
      const event = pendingTouchEvent.current;
      if (!event || !canvas || event.touches.length === 0) return;

      const now = Date.now();
      if (now - lastSendTime.current < 16) return; // 最小間隔

      pendingTouchEvent.current = null;
      lastSendTime.current = now;
      
      const rect = canvas.getBoundingClientRect();
      const touch = event.touches[0];
      const mouseEvent = touchToMouseEvent(touch);
      
      const currentPlayer = gameState.pls?.find((p) => p[0] === playerId);
      const playerPosition = currentPlayer
        ? getPlayer(currentPlayer)?.cel?.c?.p
        : undefined;

      if (!playerPosition) return;

      // カメラズーム設定を取得
      const gameZoomScale = gameConfig.cameraZoomScale || 1.0;

      const worldCoords = convertMouseToWorldCoords(mouseEvent, rect, playerPosition, gameZoomScale);
      if (!worldCoords) return;
      const { worldX, worldY } = worldCoords;

      // コアからタッチ位置への方向ベクトルを計算
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

        onTouchMove(accelerationX, accelerationY);
      } else {
        onTouchMove(0, 0);
      }

      if (animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
        animationFrameId.current = null;
      }
    };

    // タッチスタート - 即座に移動開始
    const handleTouchStart = (event: TouchEvent) => {
      event.preventDefault();
      if (event.touches.length !== 1) return;

      isTouching.current = true;
      pendingTouchEvent.current = event;
      
      // 即座に移動処理開始
      processTouchEvent();
    };

    // タッチエンド - 移動停止のみ
    const handleTouchEnd = (event: TouchEvent) => {
      event.preventDefault();
      
      isTouching.current = false;
      
      // 移動を停止
      onTouchMove(0, 0);
    };

    canvas.addEventListener("touchstart", handleTouchStart, { passive: false });
    canvas.addEventListener("touchmove", handleTouchMove, { passive: false });
    canvas.addEventListener("touchend", handleTouchEnd, { passive: false });
    canvas.addEventListener("touchcancel", handleTouchEnd, { passive: false });

    return () => {
      canvas.removeEventListener("touchstart", handleTouchStart);
      canvas.removeEventListener("touchmove", handleTouchMove);
      canvas.removeEventListener("touchend", handleTouchEnd);
      canvas.removeEventListener("touchcancel", handleTouchEnd);
      
      if (animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
      }
    };
  }, [gameState, playerId, gameConfig.cameraZoomScale, onTouchMove, onTouchClick]);
};