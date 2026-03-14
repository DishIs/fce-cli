package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DishIs/fce-cli/internal/config"
)

const baseURL = "https://api2.freecustom.email/v1"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func New() (*Client, error) {
	key, err := config.LoadAPIKey()
	if err != nil {
		return nil, err
	}
	return &Client{
		apiKey: key,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (c *Client) request(method, path string, body interface{}) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "fce-cli/1.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

func (c *Client) get(path string) (map[string]interface{}, error) {
	data, status, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return parseResponse(data, status)
}

func (c *Client) post(path string, body interface{}) (map[string]interface{}, error) {
	data, status, err := c.request("POST", path, body)
	if err != nil {
		return nil, err
	}
	return parseResponse(data, status)
}

func (c *Client) delete(path string) (map[string]interface{}, error) {
	data, status, err := c.request("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	return parseResponse(data, status)
}

func parseResponse(data []byte, status int) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response")
	}
	if status >= 400 {
		msg, _ := result["message"].(string)
		errCode, _ := result["error"].(string)
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", status)
		}
		if errCode != "" {
			return nil, fmt.Errorf("[%s] %s", errCode, msg)
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return result, nil
}

// ── API Methods ───────────────────────────────────────────────────────────────

func (c *Client) GetMe() (map[string]interface{}, error) {
	return c.get("/me")
}

func (c *Client) GetUsage() (map[string]interface{}, error) {
	return c.get("/usage")
}

func (c *Client) ListInboxes() ([]interface{}, error) {
	result, err := c.get("/inboxes")
	if err != nil {
		return nil, err
	}
	data, _ := result["data"].([]interface{})
	return data, nil
}

func (c *Client) RegisterInbox(inbox string) (map[string]interface{}, error) {
	return c.post("/inboxes", map[string]string{"inbox": inbox})
}

func (c *Client) UnregisterInbox(inbox string) (map[string]interface{}, error) {
	return c.delete("/inboxes/" + inbox)
}

func (c *Client) ListMessages(inbox string) ([]interface{}, error) {
	result, err := c.get("/inboxes/" + inbox + "/messages")
	if err != nil {
		return nil, err
	}
	data, _ := result["data"].([]interface{})
	return data, nil
}

func (c *Client) GetOTP(inbox string) (map[string]interface{}, error) {
	return c.get("/inboxes/" + inbox + "/otp")
}

func (c *Client) ListDomains() ([]interface{}, error) {
	result, err := c.get("/domains")
	if err != nil {
		return nil, err
	}
	data, _ := result["data"].([]interface{})
	return data, nil
}

func (c *Client) GetAPIKey() string {
	return c.apiKey
}

// ── Plan helpers ──────────────────────────────────────────────────────────────

type PlanLevel int

const (
	PlanFree       PlanLevel = 0
	PlanDeveloper  PlanLevel = 1
	PlanStartup    PlanLevel = 2
	PlanGrowth     PlanLevel = 3
	PlanEnterprise PlanLevel = 4
)

var planLevels = map[string]PlanLevel{
	"free":       PlanFree,
	"developer":  PlanDeveloper,
	"startup":    PlanStartup,
	"growth":     PlanGrowth,
	"enterprise": PlanEnterprise,
}

func PlanLevelFor(plan string) PlanLevel {
	if l, ok := planLevels[plan]; ok {
		return l
	}
	return PlanFree
}

func HasPlan(userPlan string, required PlanLevel) bool {
	return PlanLevelFor(userPlan) >= required
}
