package proxy

import "testing"

func TestSingBoxRegisterBridgeStoresNewBridge(t *testing.T) {
	manager := &SingBoxManager{
		Bridges: make(map[string]*SingBoxBridge),
	}
	bridge := &SingBoxBridge{
		NodeKey: "node-a",
		Port:    21001,
		Running: true,
	}

	socksURL, reused := manager.registerBridge("node-a", bridge)
	if reused {
		t.Fatalf("expected new bridge registration, got reused with %q", socksURL)
	}
	if socksURL != "" {
		t.Fatalf("expected empty socksURL for new bridge registration, got %q", socksURL)
	}
	if manager.Bridges["node-a"] != bridge {
		t.Fatalf("bridge was not stored in manager")
	}
}

func TestSingBoxRegisterBridgeIgnoresSamePointer(t *testing.T) {
	manager := &SingBoxManager{
		Bridges: make(map[string]*SingBoxBridge),
	}
	bridge := &SingBoxBridge{
		NodeKey: "node-a",
		Port:    21001,
		Running: true,
	}
	manager.Bridges["node-a"] = bridge

	socksURL, reused := manager.registerBridge("node-a", bridge)
	if reused {
		t.Fatalf("same bridge pointer must not be treated as duplicate, got reused with %q", socksURL)
	}
	if socksURL != "" {
		t.Fatalf("expected empty socksURL when registering same pointer, got %q", socksURL)
	}
	if manager.Bridges["node-a"] != bridge {
		t.Fatalf("bridge mapping changed unexpectedly")
	}
	if bridge.Stopping {
		t.Fatalf("same bridge pointer should not be marked as stopping")
	}
}

func TestBuildSingBoxOutboundSupportsAnyTLSURI(t *testing.T) {
	outbound, err := BuildSingBoxOutbound("anytls://secret@example.com:443?sni=cdn.example.com&insecure=1&alpn=h2,http/1.1&client-fingerprint=chrome")
	if err != nil {
		t.Fatalf("expected anytls uri to parse, got error: %v", err)
	}

	if outbound["type"] != "anytls" {
		t.Fatalf("expected outbound type anytls, got %#v", outbound["type"])
	}
	if outbound["server"] != "example.com" {
		t.Fatalf("expected server example.com, got %#v", outbound["server"])
	}
	if outbound["server_port"] != 443 {
		t.Fatalf("expected server_port 443, got %#v", outbound["server_port"])
	}

	tls, ok := outbound["tls"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tls map, got %#v", outbound["tls"])
	}
	if tls["server_name"] != "cdn.example.com" {
		t.Fatalf("expected server_name cdn.example.com, got %#v", tls["server_name"])
	}
	if tls["insecure"] != true {
		t.Fatalf("expected insecure true, got %#v", tls["insecure"])
	}
}

func TestBuildSingBoxOutboundSupportsAnyTLSClashYAML(t *testing.T) {
	node := `
proxies:
  - name: anytls-test
    type: anytls
    server: hk.example.com
    port: 8443
    password: secret
    sni: edge.example.com
    skip-cert-verify: true
    client-fingerprint: chrome
    alpn:
      - h2
      - http/1.1
`

	outbound, err := BuildSingBoxOutbound(node)
	if err != nil {
		t.Fatalf("expected clash anytls to parse, got error: %v", err)
	}

	if outbound["type"] != "anytls" {
		t.Fatalf("expected outbound type anytls, got %#v", outbound["type"])
	}
	if outbound["server"] != "hk.example.com" {
		t.Fatalf("expected server hk.example.com, got %#v", outbound["server"])
	}
	if outbound["password"] != "secret" {
		t.Fatalf("expected password secret, got %#v", outbound["password"])
	}
}
