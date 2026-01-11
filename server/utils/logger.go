package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type Logger struct {
	level      LogLevel
	structured bool
}

var logger = &Logger{
	level:      INFO,
	structured: os.Getenv("LOG_FORMAT") == "json",
}

func SetLogLevel(level LogLevel) {
	logger.level = level
}

func (l *Logger) log(level LogLevel, levelStr string, message string, data map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     levelStr,
		Message:   message,
		Data:      data,
	}

	if l.structured {
		jsonData, _ := json.Marshal(entry)
		fmt.Println(string(jsonData))
	} else {
		// ログレベルに応じた色とアイコンを設定
		var colorCode string
		var icon string
		switch level {
		case DEBUG:
			colorCode = "\033[36m" // Cyan
			icon = "🔍"
		case INFO:
			colorCode = "" // No color
			icon = "✅"
		case WARN:
			colorCode = "\033[33m" // Yellow
			icon = "⚠️"
		case ERROR:
			colorCode = "\033[31m" // Red
			icon = "❌"
		case FATAL:
			colorCode = "\033[35m" // Magenta
			icon = "💀"
		}

		// イベントタイプに応じた追加アイコンと色
		eventIcon := ""
		eventColor := ""
		if event, ok := data["event"].(string); ok {
			switch event {
			case "server_start":
				eventIcon = " 🚀"
			case "game_start":
				eventIcon = " 🎮"
			case "game_end":
				eventIcon = " 🏁"
			case "connect":
				eventIcon = " 👤"
				eventColor = "\033[32m" // Green
			case "npc_joined":
				eventIcon = " 👤"
			case "disconnect":
				eventIcon = " 👋"
				eventColor = "\033[35m" // Magenta
			case "player_destroyed":
				eventIcon = " 💥"
			case "collision", "satellite_collision", "core_satellite_collision", "projectile_hit_core":
				eventIcon = " 💥"
			case "satellite_eject":
				eventIcon = " 🚀"
			case "npc_batch_add":
				eventIcon = " 🤖"
			case "panic_recovery":
				eventIcon = " 🚨"
			case "websocket_error":
				eventIcon = " 📡"
			case "performance":
				eventIcon = " ⏱️"
			}
		}

		resetCode := "\033[0m"
		timeStr := time.Now().Format("15:04:05.000")

		// eventColorがあればそれを優先
		if eventColor != "" {
			colorCode = eventColor
		}

		logStr := ""
		if colorCode != "" {
			logStr = fmt.Sprintf("%s%s [%s] %-5s%s: %s%s",
				colorCode, icon, timeStr, levelStr, eventIcon, message, resetCode)
		} else {
			logStr = fmt.Sprintf("%s [%s] %-5s%s: %s",
				icon, timeStr, levelStr, eventIcon, message)
		}

		// データを見やすく整形
		if len(data) > 0 {
			// 重要な情報を先頭に表示
			importantFields := []string{"player_name", "game_id", "error", "duration_ms"}
			var parts []string

			// 重要なフィールドを優先的に表示
			for _, field := range importantFields {
				if val, ok := data[field]; ok {
					parts = append(parts, fmt.Sprintf("%s=%v", field, val))
					delete(data, field) // 後で重複表示しないように削除
				}
			}

			// 残りのフィールドを表示
			for k, v := range data {
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}

			if len(parts) > 0 {
				if colorCode != "" {
					logStr += fmt.Sprintf(" %s[%s]%s", colorCode, joinStrings(parts, ", "), resetCode)
				} else {
					logStr += fmt.Sprintf(" [%s]", joinStrings(parts, ", "))
				}
			}
		}

		if level == FATAL {
			log.Fatal(logStr)
		} else {
			log.Println(logStr)
		}
	}
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func Debug(message string, data map[string]interface{}) {
	logger.log(DEBUG, "DEBUG", message, data)
}

func Info(message string, data map[string]interface{}) {
	logger.log(INFO, "INFO", message, data)
}

func Warn(message string, data map[string]interface{}) {
	logger.log(WARN, "WARN", message, data)
}

func Error(message string, data map[string]interface{}) {
	logger.log(ERROR, "ERROR", message, data)
}

func Fatal(message string, data map[string]interface{}) {
	logger.log(FATAL, "FATAL", message, data)
}

// メトリクス用のヘルパー関数
func LogServerStart(port string) {
	Info("Server started", map[string]interface{}{
		"port":  port,
		"event": "server_start",
		"pid":   os.Getpid(),
	})
}

func LogConnectionEvent(eventType string, playerID string, playerName string, isNPC bool) {
	Info("Connection event", map[string]interface{}{
		"event":       eventType,
		"player_id":   playerID,
		"player_name": playerName,
		"is_npc":      isNPC,
		"metric":      "connection",
	})
}

func LogGameSessionEvent(eventType string, gameID string, humanPlayers int, totalPlayers int, duration time.Duration) {
	Info("Game session event", map[string]interface{}{
		"event":         eventType,
		"game_id":       gameID,
		"human_players": humanPlayers,
		"total_players": totalPlayers,
		"duration_ms":   duration.Milliseconds(),
		"metric":        "game_session",
	})
}

func LogGameMetrics(gameID string, frameCount int64, playerCount int, droppedSatellites int) {
	Debug("Game metrics", map[string]interface{}{
		"game_id":            gameID,
		"frame_count":        frameCount,
		"player_count":       playerCount,
		"dropped_satellites": droppedSatellites,
		"metric":             "game_state",
	})
}

func LogPanicRecovery(location string, gameID string, err interface{}) {
	Error("Panic recovered", map[string]interface{}{
		"location": location,
		"game_id":  gameID,
		"error":    fmt.Sprintf("%v", err),
		"metric":   "panic_recovery",
		"severity": "critical",
	})
}

func LogWebSocketError(playerID string, action string, err error) {
	Warn("WebSocket error", map[string]interface{}{
		"player_id": playerID,
		"action":    action,
		"error":     err.Error(),
		"metric":    "websocket_error",
	})
}

func LogPerformanceWarning(component string, duration time.Duration, threshold time.Duration) {
	if duration > threshold {
		Warn("Performance threshold exceeded", map[string]interface{}{
			"component":    component,
			"duration_ms":  duration.Milliseconds(),
			"threshold_ms": threshold.Milliseconds(),
			"metric":       "performance",
			"event":        "performance",
		})
	}
}

// ゲーム内イベント用の特化したログ関数
func LogGameEvent(eventType string, gameID string, details map[string]interface{}) {
	data := map[string]interface{}{
		"event":   eventType,
		"game_id": gameID,
		"metric":  "game_event",
	}
	// 詳細情報をマージ
	for k, v := range details {
		data[k] = v
	}
	Info("Game event", data)
}

// プレイヤーアクション用のログ
func LogPlayerAction(action string, playerID string, playerName string, details map[string]interface{}) {
	data := map[string]interface{}{
		"event":       action,
		"player_id":   playerID,
		"player_name": playerName,
		"metric":      "player_action",
	}
	for k, v := range details {
		data[k] = v
	}
	Info("Player action", data)
}
