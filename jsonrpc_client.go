package godoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      interface{} `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{}   `json:"id"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONRPCClient handles JSON-RPC communication with Odoo
type JSONRPCClient struct {
	baseURL    string
	httpClient *http.Client
	sessionID  string
	uid        int64
	context    map[string]interface{}
}

// NewJSONRPCClient creates a new JSON-RPC client
func NewJSONRPCClient(baseURL string, pool *Pool, timeoutSeconds int) *JSONRPCClient {
	transport := &http.Transport{}

	if pool != nil {
		transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				conn, err := pool.Get(ctx, addr)
				if err != nil {
					return nil, err
				}
				return conn, nil
			},
			MaxIdleConns:        pool.maximalIdle,
			MaxIdleConnsPerHost: pool.maxHosts,
			IdleConnTimeout:     0,
		}
	}

	// Default timeout to 30 seconds if not specified or invalid
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	return &JSONRPCClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(timeoutSeconds) * time.Second,
		},
		context: make(map[string]interface{}),
	}
}

// Call makes a JSON-RPC method call
func (c *JSONRPCClient) Call(method string, params interface{}) (*JSONRPCResponse, error) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      time.Now().UnixNano(),
	}

	// Add session context to params if available
	if c.sessionID != "" && len(c.context) > 0 {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if _, exists := paramsMap["context"]; !exists {
				paramsMap["context"] = c.context
			}
		}
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/jsonrpc", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	if rpcResp.Error != nil {
		// Include more detailed error information for debugging
		errorMsg := fmt.Sprintf("JSON-RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
		if rpcResp.Error.Data != nil {
			if dataStr, err := json.Marshal(rpcResp.Error.Data); err == nil {
				errorMsg = fmt.Sprintf("%s, data: %s", errorMsg, string(dataStr))
			}
		}
		return nil, fmt.Errorf(errorMsg)
	}

	return &rpcResp, nil
}

// Authenticate performs JSON-RPC authentication
func (c *JSONRPCClient) Authenticate(db, login, password string) error {
	// For Odoo JSON-RPC authentication, we need to call the 'jsonrpc' method
	// with service='common', method='authenticate', and args=[db, login, password, {}]
	params := map[string]interface{}{
		"service": "common",
		"method":  "authenticate",
		"args": []interface{}{
			db,
			login,
			password,
			map[string]interface{}{}, // empty context for authentication
		},
	}

	resp, err := c.Call("jsonrpc", params)
	if err != nil {
		return fmt.Errorf("authentication call failed: %w", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		// For JSON-RPC authentication, the result might be directly the UID (int64)
		if uid, ok := resp.Result.(float64); ok {
			c.uid = int64(uid)
			// Set up context for subsequent calls
			c.context = map[string]interface{}{
				"lang": "en_US",
				"tz":   "UTC",
				"uid":  c.uid,
			}
			return nil
		}
		return fmt.Errorf("invalid authentication response format, got: %T", resp.Result)
	}

	// Check if session_id is present (for newer Odoo versions)
	if sessionIDValue, ok := result["session_id"].(string); ok {
		c.sessionID = sessionIDValue
	}

	// Get UID from result
	if uidValue, ok := result["uid"].(float64); ok {
		c.uid = int64(uidValue)
	} else if uidValue, ok := resp.Result.(float64); ok {
		// Some Odoo versions return UID directly
		c.uid = int64(uidValue)
	} else {
		return fmt.Errorf("uid not found in response")
	}

	// Set up context for subsequent calls
	c.context = map[string]interface{}{
		"lang": "en_US",
		"tz":   "UTC",
		"uid":  c.uid,
	}

	return nil
}

// ExecuteKw executes a method on an Odoo model
func (c *JSONRPCClient) ExecuteKw(db string, uid int64, password string, model string, method string, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	params := map[string]interface{}{
		"service": "object",
		"method":  "execute_kw",
		"args": []interface{}{
			db,
			uid,
			password,
			model,
			method,
			args,
			kwargs,
		},
	}

	resp, err := c.Call("call", params)
	if err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// GetSessionID returns the current session ID
func (c *JSONRPCClient) GetSessionID() string {
	return c.sessionID
}

// GetUID returns the authenticated user ID
func (c *JSONRPCClient) GetUID() int64 {
	return c.uid
}

// Close closes the JSON-RPC client
func (c *JSONRPCClient) Close() {
	// JSON-RPC client doesn't maintain persistent connections that need closing
	// The underlying HTTP client will be garbage collected
}
