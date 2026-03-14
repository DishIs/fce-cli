package ws

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DishIs/fce-cli/internal/display"
	"github.com/gorilla/websocket"
)

const wsBaseURL = "wss://api2.freecustom.email/v1/ws"

type emailEvent struct {
	Type             string `json:"type"`
	ID               string `json:"id"`
	From             string `json:"from"`
	To               string `json:"to"`
	Subject          string `json:"subject"`
	Date             string `json:"date"`
	OTP              string `json:"otp"`
	VerificationLink string `json:"verificationLink"`
	HasAttachment    bool   `json:"hasAttachment"`
	Plan             string `json:"plan"`
	// Connected event fields
	SubscribedInboxes []string `json:"subscribed_inboxes"`
	ConnectionCount   int      `json:"connection_count"`
}

// Watch opens a WebSocket connection and streams emails to the terminal.
// mailbox can be "" to watch all registered inboxes.
func Watch(apiKey, mailbox string) error {
	u, _ := url.Parse(wsBaseURL)
	q := u.Query()
	q.Set("api_key", apiKey)
	if mailbox != "" {
		q.Set("mailbox", mailbox)
	}
	u.RawQuery = q.Encode()

	// Connect
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("could not connect to WebSocket: %w", err)
	}
	defer conn.Close()

	// Graceful shutdown on SIGINT / SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println()
		display.Info("Disconnected.")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
		os.Exit(0)
	}()

	// Ping loop to keep alive
	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			msg, _ := json.Marshal(map[string]string{"type": "ping"})
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()

	// Read loop
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			return fmt.Errorf("connection error: %w", err)
		}

		var event emailEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			continue
		}

		switch event.Type {
		case "connected":
			inboxList := mailbox
			if inboxList == "" {
				inboxList = fmt.Sprintf("%d inbox(es)", len(event.SubscribedInboxes))
			}
			display.Success(fmt.Sprintf("Watching %s  %s",
				inboxList,
				display.PlanBadge(event.Plan),
			))
			display.Info("Waiting for emails… (press Ctrl+C to stop)")

		case "pong":
			// keepalive — ignore

		case "error":
			// plan gate or connection error
			var errEvent struct {
				Type    string `json:"type"`
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			json.Unmarshal(raw, &errEvent)
			display.Error(errEvent.Message)
			return nil

		default:
			// Email event
			otp  := safeStr(event.OTP)
			link := safeStr(event.VerificationLink)
			ts   := formatTimestamp(event.Date)
			display.EmailEvent(event.From, event.Subject, otp, link, ts)
		}
	}
}

func safeStr(s string) string {
	if s == "null" || s == "__DETECTED__" || s == "__UPGRADE_REQUIRED__" {
		return ""
	}
	return s
}

func formatTimestamp(t string) string {
	if t == "" {
		return time.Now().Format("15:04:05")
	}
	parsed, err := time.Parse(time.RFC3339Nano, t)
	if err != nil {
		return t
	}
	return parsed.Local().Format("15:04:05")
}
