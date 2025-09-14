import React from 'react';
import './TouchControls.css';

interface TouchControlsProps {
  onDirectionChange: (direction: string) => void;
}

const TouchControls: React.FC<TouchControlsProps> = ({ onDirectionChange }) => {
  return (
    <div className="touch-controls">
      <button 
        className="control-button up"
        onTouchStart={() => onDirectionChange('UP')}
        onClick={() => onDirectionChange('UP')}
      >
        ▲
      </button>
      <div className="control-row">
        <button 
          className="control-button left"
          onTouchStart={() => onDirectionChange('LEFT')}
          onClick={() => onDirectionChange('LEFT')}
        >
          ◀
        </button>
        <div className="control-center"></div>
        <button 
          className="control-button right"
          onTouchStart={() => onDirectionChange('RIGHT')}
          onClick={() => onDirectionChange('RIGHT')}
        >
          ▶
        </button>
      </div>
      <button 
        className="control-button down"
        onTouchStart={() => onDirectionChange('DOWN')}
        onClick={() => onDirectionChange('DOWN')}
      >
        ▼
      </button>
    </div>
  );
};

export default TouchControls;