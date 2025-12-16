package mqtt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/levonbragg/go-powercontrol/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

// MessageCallback is called when a message is received
type MessageCallback func(topic string, payload string)

// ConnectionCallback is called when connection status changes
type ConnectionCallback func(connected bool)

// Client wraps the MQTT client with auto-reconnect functionality
type Client struct {
	client             mqtt.Client
	connected          bool
	mu                 sync.RWMutex
	messageCallback    MessageCallback
	connectionCallback ConnectionCallback
	ctx                context.Context
	cancel             context.CancelFunc
}

// NewClient creates a new MQTT client
func NewClient() *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		ctx:    ctx,
		cancel: cancel,
	}
}

// SetMessageCallback sets the callback for received messages
func (c *Client) SetMessageCallback(callback MessageCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageCallback = callback
}

// SetConnectionCallback sets the callback for connection status changes
func (c *Client) SetConnectionCallback(callback ConnectionCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connectionCallback = callback
}

// Connect establishes connection to the MQTT broker
func (c *Client) Connect(cfg *config.Config) error {
	// Validate config
	if cfg.MQTTServer == "" {
		return fmt.Errorf("MQTT server not configured")
	}

	// Get decrypted password
	password, err := cfg.GetPassword()
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}

	// Generate client ID
	clientID := "go-powercontrol-" + uuid.New().String()

	// Build broker URL
	brokerURL := fmt.Sprintf("tcp://%s:%d", cfg.MQTTServer, cfg.ServerPort)

	// Configure MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(password)
	opts.SetKeepAlive(5 * time.Second)
	opts.SetPingTimeout(20 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetCleanSession(true)

	// Set connection callbacks
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		c.mu.Lock()
		c.connected = true
		callback := c.connectionCallback
		c.mu.Unlock()

		if callback != nil {
			callback(true)
		}

		// Resubscribe on reconnect
		if cfg.SubscribeString != "" {
			c.Subscribe(cfg.SubscribeString)
		}
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		c.mu.Lock()
		c.connected = false
		callback := c.connectionCallback
		c.mu.Unlock()

		if callback != nil {
			callback(false)
		}
	})

	// Create and connect client
	c.client = mqtt.NewClient(opts)
	token := c.client.Connect()

	// Wait for connection with timeout
	if !token.WaitTimeout(20 * time.Second) {
		return fmt.Errorf("connection timeout")
	}

	if err := token.Error(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	return nil
}

// Subscribe subscribes to a topic
func (c *Client) Subscribe(topic string) error {
	if c.client == nil {
		return fmt.Errorf("client not initialized")
	}

	// Set message handler
	token := c.client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		c.mu.RLock()
		callback := c.messageCallback
		c.mu.RUnlock()

		if callback != nil {
			callback(msg.Topic(), string(msg.Payload()))
		}
	})

	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("subscribe timeout")
	}

	if err := token.Error(); err != nil {
		return fmt.Errorf("subscribe failed: %w", err)
	}

	return nil
}

// Publish publishes a message to a topic
func (c *Client) Publish(topic string, payload string) error {
	if c.client == nil {
		return fmt.Errorf("client not initialized")
	}

	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()

	if !connected {
		return fmt.Errorf("not connected to broker")
	}

	token := c.client.Publish(topic, 0, false, payload)

	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("publish timeout")
	}

	if err := token.Error(); err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	return nil
}

// IsConnected returns the current connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Disconnect disconnects from the MQTT broker
func (c *Client) Disconnect() {
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(250)
	}

	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()

	c.cancel()
}
