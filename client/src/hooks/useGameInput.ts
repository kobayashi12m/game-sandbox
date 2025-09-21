import { useEffect, useCallback } from 'react';
import type { Direction } from '../types';

interface UseGameInputProps {
  onDirectionChange: (direction: Direction) => void;
  isEnabled: boolean;
}

export const useGameInput = ({ onDirectionChange, isEnabled }: UseGameInputProps) => {
  // キーボード入力の処理
  useEffect(() => {
    if (!isEnabled) return;

    const handleKeyPress = (e: KeyboardEvent) => {
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

      if (direction) {
        e.preventDefault(); // デフォルトの動作を防ぐ
        onDirectionChange(direction);
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [onDirectionChange, isEnabled]);

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