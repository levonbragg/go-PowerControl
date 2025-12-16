package app

import (
	"context"
	"fmt"
	"log"

	"github.com/levonbragg/go-powercontrol/config"
	"github.com/levonbragg/go-powercontrol/models"
	"github.com/levonbragg/go-powercontrol/mqtt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx         context.Context
	mqttClient  *mqtt.Client
	deviceStore *models.DeviceStore
	messageLog  *models.MessageLog
	config      *config.Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		mqttClient:  mqtt.NewClient(),
		deviceStore: models.NewDeviceStore(),
		messageLog:  models.NewMessageLog(1000),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		cfg = config.DefaultConfig()
	}
	a.config = cfg

	// Set up MQTT callbacks
	a.mqttClient.SetMessageCallback(a.handleMQTTMessage)
	a.mqttClient.SetConnectionCallback(a.handleConnectionStatus)

	// Auto-connect if config is valid
	if !cfg.IsEmpty() {
		go func() {
			if err := a.connectMQTT(); err != nil {
				log.Printf("Auto-connect failed: %v", err)
			}
		}()
	}
}

// Shutdown is called when the app is closing
func (a *App) Shutdown(ctx context.Context) {
	a.mqttClient.Disconnect()
}

// connectMQTT connects to the MQTT broker
func (a *App) connectMQTT() error {
	if err := a.mqttClient.Connect(a.config); err != nil {
		return err
	}

	// Subscribe to the configured topic
	if err := a.mqttClient.Subscribe(a.config.SubscribeString); err != nil {
		return err
	}

	return nil
}

// handleMQTTMessage processes incoming MQTT messages
func (a *App) handleMQTTMessage(topic string, payload string) {
	// Log the message
	a.messageLog.AddMessage(models.MessageReceived, topic, payload)

	// Emit event to frontend
	runtime.EventsEmit(a.ctx, "message:new", map[string]interface{}{
		"direction": "Recv",
		"topic":     topic,
		"payload":   payload,
	})

	// Parse topic to extract device and outlet
	device, outlet, err := mqtt.ParseTopic(topic)
	if err != nil {
		log.Printf("Failed to parse topic %s: %v", topic, err)
		return
	}

	// Parse payload to get status
	status := mqtt.ParsePayload(payload)

	// Update device store
	deviceOutlet := models.DeviceOutlet{
		DeviceName:   device,
		OutletNumber: outlet,
		Status:       status,
	}
	a.deviceStore.Add(deviceOutlet)

	// Emit device update event to frontend
	runtime.EventsEmit(a.ctx, "device:update", deviceOutlet)
}

// handleConnectionStatus processes connection status changes
func (a *App) handleConnectionStatus(connected bool) {
	// Emit connection status event to frontend
	runtime.EventsEmit(a.ctx, "connection:status", connected)
}

// GetConnectionStatus returns the current MQTT connection status
func (a *App) GetConnectionStatus() bool {
	return a.mqttClient.IsConnected()
}

// GetDevices returns all devices
func (a *App) GetDevices() []models.DeviceOutlet {
	return a.deviceStore.GetAll()
}

// SearchDevices returns filtered devices based on search text
func (a *App) SearchDevices(searchText string) []models.DeviceOutlet {
	return a.deviceStore.Filter(searchText)
}

// GetMessages returns all logged messages
func (a *App) GetMessages() []models.MQTTMessage {
	return a.messageLog.GetAll()
}

// SaveSettings saves the configuration and reconnects if necessary
func (a *App) SaveSettings(username, password, server string, port int, subscribeString string) error {
	// Create new config
	cfg := &config.Config{
		Username:        username,
		MQTTServer:      server,
		ServerPort:      port,
		SubscribeString: subscribeString,
	}

	// Encrypt and set password
	if err := cfg.SetPassword(password); err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save to disk
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Update current config
	a.config = cfg

	// Disconnect and reconnect with new settings
	a.mqttClient.Disconnect()

	// Clear devices and messages on reconnect
	a.deviceStore.Clear()

	// Connect with new config
	if err := a.connectMQTT(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	return nil
}

// SendCommand publishes a command to turn an outlet on or off
func (a *App) SendCommand(deviceName, outletNumber, state string) error {
	// Build command topic
	topic := mqtt.MakeCommandTopic(deviceName, outletNumber)

	// Convert state to payload
	payload := mqtt.StatusToPayload(state)

	// Publish
	if err := a.mqttClient.Publish(topic, payload); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	// Log the sent message
	a.messageLog.AddMessage(models.MessageSent, topic, payload)

	// Emit event to frontend
	runtime.EventsEmit(a.ctx, "message:new", map[string]interface{}{
		"direction": "Send",
		"topic":     topic,
		"payload":   payload,
	})

	return nil
}

// Disconnect disconnects from the MQTT broker
func (a *App) Disconnect() error {
	a.mqttClient.Disconnect()
	return nil
}

// ClearLog clears the message log
func (a *App) ClearLog() {
	a.messageLog.Clear()
	runtime.EventsEmit(a.ctx, "log:cleared")
}

// GetConfig returns the current configuration (without password)
func (a *App) GetConfig() map[string]interface{} {
	if a.config == nil {
		return map[string]interface{}{
			"username":        "",
			"mqttServer":      "",
			"serverPort":      1883,
			"subscribeString": "power/#",
		}
	}

	return map[string]interface{}{
		"username":        a.config.Username,
		"mqttServer":      a.config.MQTTServer,
		"serverPort":      a.config.ServerPort,
		"subscribeString": a.config.SubscribeString,
	}
}

// IsConfigEmpty returns true if the configuration is not set up
func (a *App) IsConfigEmpty() bool {
	return a.config == nil || a.config.IsEmpty()
}
