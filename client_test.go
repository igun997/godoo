package godoo

import (
	"testing"
)

func TestClientProtocolMethods(t *testing.T) {
	pool := NewPool(5, 2, 30)

	// Test JSON-RPC (default) - don't authenticate to avoid network calls
	cfgJSONRPC := &ClientConfig{
		Database: "test_db",
		Admin:    "admin",
		Password: "password",
		URL:      "https://test.odoo.com",
		Pool:     pool,
		Timeout:  30,
		// Protocol: ProtocolJSONRPC // This is now the default
	}

	clientJSONRPC := &Client{
		cfg:      cfgJSONRPC,
		jsonrpc:  NewJSONRPCClient(cfgJSONRPC.URL, cfgJSONRPC.Pool, 30),
		auth:     false,
		protocol: ProtocolJSONRPC,
	}

	if clientJSONRPC.GetProtocol() != ProtocolJSONRPC {
		t.Errorf("Expected protocol %s, got %s", ProtocolJSONRPC, clientJSONRPC.GetProtocol())
	}

	if !clientJSONRPC.IsJSONRPC() {
		t.Error("Expected IsJSONRPC() to return true")
	}

	if clientJSONRPC.IsXMLRPC() {
		t.Error("Expected IsXMLRPC() to return false")
	}

	// Test XML-RPC (explicit) - don't authenticate to avoid network calls
	cfgXMLRPC := &ClientConfig{
		Database: "test_db",
		Admin:    "admin",
		Password: "password",
		URL:      "https://test.odoo.com",
		Pool:     pool,
		Protocol: ProtocolXMLRPC,
	}

	clientXMLRPC := &Client{
		cfg:      cfgXMLRPC,
		auth:     false,
		protocol: ProtocolXMLRPC,
	}

	if clientXMLRPC.GetProtocol() != ProtocolXMLRPC {
		t.Errorf("Expected protocol %s, got %s", ProtocolXMLRPC, clientXMLRPC.GetProtocol())
	}

	if !clientXMLRPC.IsXMLRPC() {
		t.Error("Expected IsXMLRPC() to return true")
	}

	if clientXMLRPC.IsJSONRPC() {
		t.Error("Expected IsJSONRPC() to return false")
	}

	clientXMLRPC.Close()
	clientJSONRPC.Close()
}

func TestClientConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *ClientConfig
		valid  bool
	}{
		{
			name: "Valid config",
			config: &ClientConfig{
				Database: "test_db",
				Admin:    "admin",
				Password: "password",
				URL:      "https://test.odoo.com",
			},
			valid: true,
		},
		{
			name: "Missing database",
			config: &ClientConfig{
				Admin:    "admin",
				Password: "password",
				URL:      "https://test.odoo.com",
			},
			valid: false,
		},
		{
			name: "Missing admin",
			config: &ClientConfig{
				Database: "test_db",
				Password: "password",
				URL:      "https://test.odoo.com",
			},
			valid: false,
		},
		{
			name: "Missing password",
			config: &ClientConfig{
				Database: "test_db",
				Admin:    "admin",
				URL:      "https://test.odoo.com",
			},
			valid: false,
		},
		{
			name: "Missing URL",
			config: &ClientConfig{
				Database: "test_db",
				Admin:    "admin",
				Password: "password",
			},
			valid: false,
		},
		{
			name: "Valid with JSON-RPC protocol",
			config: &ClientConfig{
				Database: "test_db",
				Admin:    "admin",
				Password: "password",
				URL:      "https://test.odoo.com",
				Protocol: ProtocolJSONRPC,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.valid()
			if result != tt.valid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.valid, result)
			}
		})
	}
}

func TestOptionsConversion(t *testing.T) {
	options := NewOptions()
	options.Limit(10).Offset(5).FetchFields("name", "email")

	// Test that Options can be converted to map[string]interface{}
	var kwargs map[string]interface{}
	if options != nil {
		kwargs = *options
	}

	if kwargs["limit"] != 10 {
		t.Errorf("Expected limit 10, got %v", kwargs["limit"])
	}

	if kwargs["offset"] != 5 {
		t.Errorf("Expected offset 5, got %v", kwargs["offset"])
	}

	fields, ok := kwargs["fields"].([]string)
	if !ok {
		t.Error("Expected fields to be []string")
	}

	if len(fields) != 2 || fields[0] != "name" || fields[1] != "email" {
		t.Errorf("Expected fields [name, email], got %v", fields)
	}
}

func TestCriteriaCreation(t *testing.T) {
	criteria := NewCriteria()
	criteria.Add("active", "=", true)
	criteria.Add("name", "ilike", "test")

	if len(*criteria) != 2 {
		t.Errorf("Expected 2 criteria, got %d", len(*criteria))
	}

	// Test first criterion
	firstCriterion := (*criteria)[0]
	if len(*firstCriterion) != 3 {
		t.Errorf("Expected criterion with 3 elements, got %d", len(*firstCriterion))
	}

	if (*firstCriterion)[0] != "active" {
		t.Errorf("Expected field 'active', got %v", (*firstCriterion)[0])
	}

	if (*firstCriterion)[1] != "=" {
		t.Errorf("Expected operator '=', got %v", (*firstCriterion)[1])
	}

	if (*firstCriterion)[2] != true {
		t.Errorf("Expected value true, got %v", (*firstCriterion)[2])
	}

	// Test argsFromCriteria function
	args := argsFromCriteria(criteria)
	if len(args) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(args))
	}

	// argsFromCriteria returns the dereferenced criteria, not a pointer
	criteriaArray, ok := args[0].(Criteria)
	if !ok {
		t.Error("Expected criteria array")
	}

	if len(criteriaArray) != 2 {
		t.Errorf("Expected 2 criteria, got %d", len(criteriaArray))
	}
}
