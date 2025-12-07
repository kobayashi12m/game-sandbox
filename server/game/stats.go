package game

import "fmt"

// addBytesSent は送信バイト数を安全に追加する
func (g *Game) addBytesSent(bytes int) {
	if bytes > 0 {
		g.mu.Lock()
		g.totalBytesSent += int64(bytes)
		g.mu.Unlock()
	}
}

// getTotalBytes は送信バイト数の合計を取得する
func (g *Game) getTotalBytes() int64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.totalBytesSent
}

// formatBytes はバイト数を読みやすい形式に変換する
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}