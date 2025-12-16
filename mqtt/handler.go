package mqtt

import (
	"fmt"
	"strings"
)

// ParseTopic extracts device name and outlet number from MQTT topic
// Expected format: power/<device-name>/outlets/<outlet-number>
// Returns device name, outlet number, and error if parsing fails
func ParseTopic(topic string) (device string, outlet string, err error) {
	parts := strings.Split(topic, "/")

	// Expected: ["power", "<device>", "outlets", "<number>"] or
	//           ["power", "<device>", "outlets", "<number>", "set"]
	if len(parts) < 4 {
		return "", "", fmt.Errorf("invalid topic format: %s", topic)
	}

	if parts[0] != "power" {
		return "", "", fmt.Errorf("topic does not start with 'power': %s", topic)
	}

	if parts[2] != "outlets" {
		return "", "", fmt.Errorf("invalid topic structure: %s", topic)
	}

	device = parts[1]
	outlet = parts[3]

	if device == "" || outlet == "" {
		return "", "", fmt.Errorf("empty device or outlet in topic: %s", topic)
	}

	return device, outlet, nil
}

// ParsePayload converts payload string to human-readable status
// "0" -> "OFF", "1" -> "ON"
func ParsePayload(payload string) string {
	payload = strings.TrimSpace(payload)
	switch payload {
	case "0":
		return "OFF"
	case "1":
		return "ON"
	default:
		return payload // Return as-is if not 0 or 1
	}
}

// MakeCommandTopic creates the command topic for a device/outlet
// Format: power/<device>/outlets/<outlet>/set
func MakeCommandTopic(device, outlet string) string {
	return fmt.Sprintf("power/%s/outlets/%s/set", device, outlet)
}

// StatusToPayload converts status string to MQTT payload
// "OFF" -> "0", "ON" -> "1"
func StatusToPayload(status string) string {
	status = strings.ToUpper(strings.TrimSpace(status))
	switch status {
	case "OFF":
		return "0"
	case "ON":
		return "1"
	default:
		return status
	}
}
