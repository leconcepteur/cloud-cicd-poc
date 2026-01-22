package middleware

import (
	"encoding/json"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	Latency    string `json:"latency"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	Error      string `json:"error,omitempty"`
}

func JSONLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			entry := LogEntry{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Level:     "info",
				Method:    c.Request().Method,
				Path:      c.Request().URL.Path,
				Status:    c.Response().Status,
				Latency:   time.Since(start).String(),
				IP:        c.RealIP(),
				UserAgent: c.Request().UserAgent(),
			}

			if err != nil {
				entry.Level = "error"
				entry.Error = err.Error()
			} else if c.Response().Status >= 400 {
				entry.Level = "warn"
			}

			jsonBytes, _ := json.Marshal(entry)
			os.Stdout.Write(jsonBytes)
			os.Stdout.Write([]byte("\n"))

			return err
		}
	}
}
