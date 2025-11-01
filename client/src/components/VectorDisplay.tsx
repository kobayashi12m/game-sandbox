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

  // 配列形式の Position [x, y] にアクセス
  const vx = velocity[0] || 0;
  const vy = velocity[1] || 0;
  const ax = acceleration[0] || 0;
  const ay = acceleration[1] || 0;

  // ベクトルの大きさを計算
  const velocityMagnitude = Math.sqrt(vx * vx + vy * vy);
  const accelerationMagnitude = Math.sqrt(ax * ax + ay * ay);

  return (
    <div className="vector-display">
      <div className="vector-info">
        <h3>ベクトル情報</h3>
        <div className="vector-item">
          <span className="vector-label">速度:</span>
          <div className="vector-values">
            <span>X: {vx.toFixed(1)}</span>
            <span>Y: {vy.toFixed(1)}</span>
            <span>
              大きさ: {velocityMagnitude.toFixed(1)} / {maxSpeed}
            </span>
          </div>
        </div>
        <div className="vector-item">
          <span className="vector-label">加速度:</span>
          <div className="vector-values">
            <span>X: {ax.toFixed(1)}</span>
            <span>Y: {ay.toFixed(1)}</span>
            <span>大きさ: {accelerationMagnitude.toFixed(1)}</span>
          </div>
        </div>
      </div>
    </div>
  );
};
