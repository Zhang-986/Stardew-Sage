package relay

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const absoluteMaxRequestBytes int64 = 2 << 20

type Config struct {
	ListenAddress   string
	UpstreamAddress string
	LANToken        string
	CertSHA256      [32]byte
	RequestTimeout  time.Duration
	MaxRequestBytes int64
	MaxInFlight     int
}

func LoadConfig(lookup func(string) (string, bool)) (Config, error) {
	get := func(key, fallback string) string {
		if value, ok := lookup(key); ok && value != "" {
			return value
		}
		return fallback
	}
	config := Config{ListenAddress: get("ECHOFARM_RELAY_ADDRESS", "127.0.0.1:18471")}

	if err := validateLoopbackAddress(config.ListenAddress); err != nil {
		return Config{}, fmt.Errorf("ECHOFARM_RELAY_ADDRESS: %w", err)
	}
	config.UpstreamAddress = get("ECHOFARM_RELAY_UPSTREAM_ADDRESS", "")
	if config.UpstreamAddress == "" {
		return Config{}, fmt.Errorf("ECHOFARM_RELAY_UPSTREAM_ADDRESS is required")
	}
	upstreamHost, _, err := net.SplitHostPort(config.UpstreamAddress)
	if err != nil || !privateHost(upstreamHost) {
		return Config{}, fmt.Errorf("ECHOFARM_RELAY_UPSTREAM_ADDRESS must be a private host and port")
	}

	config.LANToken = get("ECHOFARM_RELAY_TOKEN", "")
	if _, err := decodeHex32(config.LANToken); err != nil {
		return Config{}, fmt.Errorf("ECHOFARM_RELAY_TOKEN must be exactly 64 hexadecimal characters")
	}
	fingerprint, err := decodeHex32(get("ECHOFARM_RELAY_CERT_SHA256", ""))
	if err != nil {
		return Config{}, fmt.Errorf("ECHOFARM_RELAY_CERT_SHA256 must be exactly 64 hexadecimal characters")
	}
	copy(config.CertSHA256[:], fingerprint)

	timeoutSeconds, err := positiveInt(get("ECHOFARM_RELAY_TIMEOUT_SECONDS", "100"), "ECHOFARM_RELAY_TIMEOUT_SECONDS")
	if err != nil {
		return Config{}, err
	}
	config.RequestTimeout = time.Duration(timeoutSeconds) * time.Second

	maxBytes, err := positiveInt64(get("ECHOFARM_RELAY_MAX_REQUEST_BYTES", strconv.FormatInt(absoluteMaxRequestBytes, 10)), "ECHOFARM_RELAY_MAX_REQUEST_BYTES")
	if err != nil || maxBytes > absoluteMaxRequestBytes {
		return Config{}, fmt.Errorf("ECHOFARM_RELAY_MAX_REQUEST_BYTES must be between 1 and %d", absoluteMaxRequestBytes)
	}
	config.MaxRequestBytes = maxBytes

	maxInFlight, err := positiveInt(get("ECHOFARM_RELAY_MAX_IN_FLIGHT", "2"), "ECHOFARM_RELAY_MAX_IN_FLIGHT")
	if err != nil || maxInFlight > 8 {
		return Config{}, fmt.Errorf("ECHOFARM_RELAY_MAX_IN_FLIGHT must be between 1 and 8")
	}
	config.MaxInFlight = maxInFlight
	return config, nil
}

func validateLoopbackAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	ip := net.ParseIP(host)
	if !strings.EqualFold(host, "localhost") && (ip == nil || !ip.IsLoopback()) {
		return fmt.Errorf("must use a loopback host")
	}
	return nil
}

func privateHost(host string) bool {
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".local") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsPrivate() || ip.IsLoopback())
}

func decodeHex32(raw string) ([]byte, error) {
	if len(raw) != 64 {
		return nil, fmt.Errorf("invalid length")
	}
	decoded, err := hex.DecodeString(raw)
	if err != nil || len(decoded) != 32 {
		return nil, fmt.Errorf("invalid hexadecimal value")
	}
	return decoded, nil
}

func positiveInt(raw, key string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}

func positiveInt64(raw, key string) (int64, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}
