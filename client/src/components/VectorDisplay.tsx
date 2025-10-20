import React from "react";
import type { Position } from "../types";

interface VectorDisplayProps {
  velocity: Position | undefined;
  acceleration: Position | undefined;
  maxSpeed: number;
}

export const VectorDisplay: React.FC<VectorDisplayProps> = ({
  velocity,
  acceleration,
  maxSpeed,
}) => {
  if (!velocity || !acceleration) return null;

  // ベクトルの大きさを計算
  const velocityMagnitude = Math.sqrt(
    velocity.x * velocity.x + velocity.y * velocity.y
  );
  const accelerationMagnitude = Math.sqrt(
    acceleration.x * acceleration.x + acceleration.y * acceleration.y
  );

  return (
    <div className="vector-display">
      <div className="vector-info">
        <h3>ベクトル情報</h3>
        <div className="vector-item">
          <span className="vector-label">速度:</span>
          <div className="vector-values">
            <span>X: {velocity.x.toFixed(1)}</span>
            <span>Y: {velocity.y.toFixed(1)}</span>
            <span>
              大きさ: {velocityMagnitude.toFixed(1)} / {maxSpeed}
            </span>
          </div>
        </div>
        <div className="vector-item">
          <span className="vector-label">加速度:</span>
          <div className="vector-values">
            <span>X: {acceleration.x.toFixed(1)}</span>
            <span>Y: {acceleration.y.toFixed(1)}</span>
            <span>大きさ: {accelerationMagnitude.toFixed(1)}</span>
          </div>
        </div>
      </div>
    </div>
  );
};
