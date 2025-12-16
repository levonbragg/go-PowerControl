package models

import (
	"strings"
	"sync"
	"time"
)

// DeviceOutlet represents a single outlet on a power device
type DeviceOutlet struct {
	DeviceName   string    `json:"deviceName"`
	OutletNumber string    `json:"outletNumber"`
	Status       string    `json:"status"` // "ON" or "OFF"
	LastUpdate   time.Time `json:"lastUpdate"`
}

// DeviceStore manages the collection of devices and outlets
type DeviceStore struct {
	mu      sync.RWMutex
	devices map[string]*DeviceOutlet // key: "deviceName:outletNumber"
}

// NewDeviceStore creates a new device store
func NewDeviceStore() *DeviceStore {
	return &DeviceStore{
		devices: make(map[string]*DeviceOutlet),
	}
}

// makeKey creates a unique key for device-outlet combination
func makeKey(deviceName, outletNumber string) string {
	return deviceName + ":" + outletNumber
}

// Add adds or updates a device outlet
func (s *DeviceStore) Add(device DeviceOutlet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device.LastUpdate = time.Now()
	key := makeKey(device.DeviceName, device.OutletNumber)
	s.devices[key] = &device
}

// Get retrieves a device outlet
func (s *DeviceStore) Get(deviceName, outletNumber string) (DeviceOutlet, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := makeKey(deviceName, outletNumber)
	device, exists := s.devices[key]
	if !exists {
		return DeviceOutlet{}, false
	}
	return *device, true
}

// GetAll returns all devices sorted by device name, then outlet number
func (s *DeviceStore) GetAll() []DeviceOutlet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]DeviceOutlet, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, *device)
	}

	// Sort by device name, then outlet number
	// Simple bubble sort for simplicity
	for i := 0; i < len(devices); i++ {
		for j := i + 1; j < len(devices); j++ {
			if devices[i].DeviceName > devices[j].DeviceName ||
				(devices[i].DeviceName == devices[j].DeviceName &&
					devices[i].OutletNumber > devices[j].OutletNumber) {
				devices[i], devices[j] = devices[j], devices[i]
			}
		}
	}

	return devices
}

// Filter returns devices matching the search text (case-insensitive)
func (s *DeviceStore) Filter(searchText string) []DeviceOutlet {
	if searchText == "" {
		return s.GetAll()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	searchText = strings.ToLower(searchText)
	filtered := make([]DeviceOutlet, 0)

	for _, device := range s.devices {
		if strings.Contains(strings.ToLower(device.DeviceName), searchText) ||
			strings.Contains(strings.ToLower(device.OutletNumber), searchText) ||
			strings.Contains(strings.ToLower(device.Status), searchText) {
			filtered = append(filtered, *device)
		}
	}

	// Sort results
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[i].DeviceName > filtered[j].DeviceName ||
				(filtered[i].DeviceName == filtered[j].DeviceName &&
					filtered[i].OutletNumber > filtered[j].OutletNumber) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	return filtered
}

// Count returns the total number of devices
func (s *DeviceStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.devices)
}

// Clear removes all devices
func (s *DeviceStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices = make(map[string]*DeviceOutlet)
}
