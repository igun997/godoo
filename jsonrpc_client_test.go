package godoo

import (
	"testing"
)

func TestProtocolType(t *testing.T) {
	tests := []struct {
		name     string
		protocol ProtocolType
		expected string
	}{
		{"XML-RPC Protocol", ProtocolXMLRPC, "xmlrpc"},
		{"JSON-RPC Protocol", ProtocolJSONRPC, "jsonrpc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.protocol) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.protocol)
			}
		})
	}
}

func TestNewJSONRPCClient(t *testing.T) {
	baseURL := "https://test.odoo.com"
	pool := NewPool(5, 2, 30)
	
	client := NewJSONRPCClient(baseURL, pool)
	
	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}
	
	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
	
	if client.sessionID != "" {
		t.Error("Expected sessionID to be empty initially")
	}
	
	if client.uid != 0 {
		t.Error("Expected uid to be 0 initially")
	}
}


func TestJSONRPCRequest(t *testing.T) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "test_method",
		Params:  map[string]interface{}{"key": "value"},
		ID:      1,
	}
	
	if request.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC version 2.0, got %s", request.JSONRPC)
	}
	
	if request.Method != "test_method" {
		t.Errorf("Expected method test_method, got %s", request.Method)
	}
	
	if request.ID != 1 {
		t.Errorf("Expected ID 1, got %v", request.ID)
	}
}

func TestJSONRPCResponse(t *testing.T) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  map[string]interface{}{"result": "success"},
		ID:      1,
	}
	
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC version 2.0, got %s", response.JSONRPC)
	}
	
	if response.Result.(map[string]interface{})["result"] != "success" {
		t.Error("Expected result to be success")
	}
	
	if response.ID != 1 {
		t.Errorf("Expected ID 1, got %v", response.ID)
	}
}