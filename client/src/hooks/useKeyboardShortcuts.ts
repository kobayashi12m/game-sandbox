import { useEffect } from 'react';
import { UI_CONFIG } from '../constants/ui';

interface UseKeyboardShortcutsProps {
  onToggleGrid: () => void;
  onToggleCulling: () => void;
}

export const useKeyboardShortcuts = ({ onToggleGrid, onToggleCulling }: UseKeyboardShortcutsProps) => {
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      // SHOW_LEFT_UIがfalseの場合は切り替えを無効化
      if (!UI_CONFIG.SHOW_LEFT_UI) return;

      if (event.key === "g" || event.key === "G") {
        onToggleGrid();
      }
      if (event.key === "c" || event.key === "C") {
        onToggleCulling();
      }
    };

    window.addEventListener("keydown", handleKeyPress);
    return () => window.removeEventListener("keydown", handleKeyPress);
  }, [onToggleGrid, onToggleCulling]);
};