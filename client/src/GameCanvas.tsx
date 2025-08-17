import React, { useRef, useEffect } from 'react';

interface Player {
  id: string;
  x: number;
  y: number;
  coreSize: number;
  guardSize: number;
}

interface GameCanvasProps {
  player: Player | null;
  otherPlayers: Player[];
  onPositionUpdate?: (x: number, y: number) => void;
}

const GameCanvas: React.FC<GameCanvasProps> = ({ player, otherPlayers, onPositionUpdate }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [mousePos, setMousePos] = React.useState({ x: 400, y: 300 });

  // マウス移動ハンドラー
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const handleMouseMove = (event: MouseEvent) => {
      const rect = canvas.getBoundingClientRect();
      const x = event.clientX - rect.left;
      const y = event.clientY - rect.top;
      
      setMousePos({ x, y });
      
      // 位置更新をコールバック
      if (onPositionUpdate) {
        onPositionUpdate(x, y);
      }
    };

    canvas.addEventListener('mousemove', handleMouseMove);

    return () => {
      canvas.removeEventListener('mousemove', handleMouseMove);
    };
  }, [onPositionUpdate]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Canvasサイズを設定
    canvas.width = 800;
    canvas.height = 600;

    // 描画関数
    const draw = () => {
      // 背景をクリア
      ctx.fillStyle = '#f0f0f0';
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // デバッグ情報を描画
      ctx.fillStyle = 'black';
      ctx.font = '14px Arial';
      ctx.textAlign = 'left';
      ctx.fillText(`他プレイヤー数: ${otherPlayers.length}`, 10, 20);

      // 他のプレイヤーを描画
      console.log('描画する他プレイヤー数:', otherPlayers.length);
      otherPlayers.forEach((otherPlayer, index) => {
        console.log(`プレイヤー${index + 1}: ${otherPlayer.id} at (${otherPlayer.x}, ${otherPlayer.y})`);
      });
      
      otherPlayers.forEach(otherPlayer => {
        // ガード（外側の円）を描画
        ctx.beginPath();
        ctx.arc(otherPlayer.x, otherPlayer.y, otherPlayer.guardSize, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(100, 100, 100, 0.3)';
        ctx.fill();
        ctx.strokeStyle = 'rgba(100, 100, 100, 0.8)';
        ctx.lineWidth = 2;
        ctx.stroke();

        // コア（内側の円）を描画
        ctx.beginPath();
        ctx.arc(otherPlayer.x, otherPlayer.y, otherPlayer.coreSize, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(150, 150, 150, 0.8)';
        ctx.fill();
        ctx.strokeStyle = 'rgba(100, 100, 100, 1)';
        ctx.lineWidth = 2;
        ctx.stroke();

        // プレイヤーIDを表示
        ctx.fillStyle = 'black';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(otherPlayer.id, otherPlayer.x, otherPlayer.y - otherPlayer.guardSize - 10);
      });

      // 自分のプレイヤーを描画（最前面）
      if (player) {
        // マウス位置にプレイヤーを描画
        const drawX = mousePos.x;
        const drawY = mousePos.y;

        // ガード（外側の円）を描画
        ctx.beginPath();
        ctx.arc(drawX, drawY, player.guardSize, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(0, 100, 255, 0.3)';
        ctx.fill();
        ctx.strokeStyle = 'rgba(0, 100, 255, 0.8)';
        ctx.lineWidth = 2;
        ctx.stroke();

        // コア（内側の円）を描画
        ctx.beginPath();
        ctx.arc(drawX, drawY, player.coreSize, 0, Math.PI * 2);
        ctx.fillStyle = 'rgba(255, 0, 0, 0.8)';
        ctx.fill();
        ctx.strokeStyle = 'rgba(200, 0, 0, 1)';
        ctx.lineWidth = 2;
        ctx.stroke();

        // プレイヤーIDを表示
        ctx.fillStyle = 'black';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(player.id, drawX, drawY - player.guardSize - 10);
      }
    };

    // 初回描画
    draw();

    // アニメーションループ
    const animate = () => {
      draw();
      requestAnimationFrame(animate);
    };

    const animationId = requestAnimationFrame(animate);

    return () => {
      cancelAnimationFrame(animationId);
    };
  }, [player, otherPlayers, mousePos]);

  return (
    <canvas
      ref={canvasRef}
      style={{
        border: '1px solid #ccc',
        display: 'block',
        margin: '20px auto',
        borderRadius: '8px',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)'
      }}
    />
  );
};

export default GameCanvas;