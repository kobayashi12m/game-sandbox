import { useEffect, useCallback, useRef } from 'react';
import type { Direction } from '../types';

interface UseGameInputProps {
  onDirectionChange: (direction: Direction) => void;
  onAccelerationChange: (x: number, y: number) => void;
  onMovementStop: () => void;
  isEnabled: boolean;
}

export const useGameInput = ({ onDirectionChange, onAccelerationChange, onMovementStop, isEnabled }: UseGameInputProps) => {
  const pressedDirections = useRef(new Set<Direction>());
  
  // 現在押されている方向から加速度ベクトルを計算
  const calculateAcceleration = (): { x: number, y: number } => {
    let x = 0, y = 0;
    
    if (pressedDirections.current.has('UP')) y -= 1;
    if (pressedDirections.current.has('DOWN')) y += 1;
    if (pressedDirections.current.has('LEFT')) x -= 1;
    if (pressedDirections.current.has('RIGHT')) x += 1;
    
    // 対角線の場合は正規化（斜め移動が速くならないように）
    if (x !== 0 && y !== 0) {
      const length = Math.sqrt(x * x + y * y);
      x /= length;
      y /= length;
    }
    
    return { x, y };
  };
  
  // キーボード入力の処理
  useEffect(() => {
    if (!isEnabled) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      let direction: Direction | null = null;
      
      switch (e.key) {
        case 'ArrowUp':
        case 'w':
        case 'W':
          direction = 'UP';
          break;
        case 'ArrowDown':
        case 's':
        case 'S':
          direction = 'DOWN';
          break;
        case 'ArrowLeft':
        case 'a':
        case 'A':
          direction = 'LEFT';
          break;
        case 'ArrowRight':
        case 'd':
        case 'D':
          direction = 'RIGHT';
          break;
      }

      if (direction && !pressedDirections.current.has(direction)) {
        e.preventDefault();
        pressedDirections.current.add(direction);
        
        // 新しい加速度ベクトルを計算して送信
        const acceleration = calculateAcceleration();
        onAccelerationChange(acceleration.x, acceleration.y);
      }
    };

    const handleKeyUp = (e: KeyboardEvent) => {
      let direction: Direction | null = null;
      
      switch (e.key) {
        case 'ArrowUp':
        case 'w':
        case 'W':
          direction = 'UP';
          break;
        case 'ArrowDown':
        case 's':
        case 'S':
          direction = 'DOWN';
          break;
        case 'ArrowLeft':
        case 'a':
        case 'A':
          direction = 'LEFT';
          break;
        case 'ArrowRight':
        case 'd':
        case 'D':
          direction = 'RIGHT';
          break;
      }

      if (direction && pressedDirections.current.has(direction)) {
        pressedDirections.current.delete(direction);
        
        // まだ押されている方向があるかチェック
        const acceleration = calculateAcceleration();
        if (acceleration.x !== 0 || acceleration.y !== 0) {
          onAccelerationChange(acceleration.x, acceleration.y);
        } else {
          onMovementStop();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
      pressedDirections.current.clear();
    };
  }, [onDirectionChange, onMovementStop, isEnabled]);

  // タッチ・マウス入力の処理（TouchControlsから呼び出される）
  const handleTouchDirection = useCallback((direction: Direction) => {
    if (isEnabled) {
      onDirectionChange(direction);
    }
  }, [onDirectionChange, isEnabled]);

  return {
    handleTouchDirection
  };
};