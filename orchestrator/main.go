package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
    CLIENT_ID       = "go-orchestrator"
    CONNECT_RETRIES = 30
    RETRY_DELAY     = 2 * time.Second
    
    // Topics
    TOPIC_SESSION_START = "hub/session/start"     // IN: From Backend API
    TOPIC_DEVICE_CMD    = "devices/%s/command"    // OUT: To specific Device
    TOPIC_DEVICE_STATUS = "devices/+/status"      // IN: Heartbeats from devices
)

// --- Data Models ---

// SessionStartPayload comes from the Python Backend
type SessionStartPayload struct {
    SessionID   string   `json:"session_id"`
    TestPlan    string   `json:"test_plan"`    // e.g., "full_suite", "smoke_test"
    TargetGroup string   `json:"target_group"` // e.g., "production_line_1", "lab_bench_A"
    BuildVer    string   `json:"build_version"`
}

// DeviceCommand is what we send to the physical hardware
type DeviceCommand struct {
    Action    string `json:"action"`      // "start_test", "stop", "reboot"
    SessionID string `json:"session_id"`  // Tag results with this ID
    Plan      string `json:"plan"`        // Which suite to run
    Config    string `json:"config_url"`  // Optional: URL to download firmware/config
}

// DeviceRegistry (In-Memory for now, replace with Redis/DB later)
type DeviceRegistry struct {
    sync.RWMutex
    // Map: GroupName -> List of DeviceIDs
    Groups map[string][]string
}

// Global Registry Instance
var registry = &DeviceRegistry{
    Groups: map[string][]string{
        // Example Mapping
        "lab_bench_A": {"device_esp32_01", "device_esp32_02"},
        "prod_line_1": {"device_arm_05", "device_arm_06", "device_arm_07"},
    },
}

// --- MQTT Helpers ---

func getBrokerURL() string {
    broker := os.Getenv("MQTT_BROKER_URL")
    if broker == "" {
        return "tcp://mosquitto:1883" 
    }
    return broker
}

func main() {
    // 1. Setup Logging & Config
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    brokerURL := getBrokerURL()
    log.Printf("[Init] Connecting to Broker: %s", brokerURL)

    // 2. Configure MQTT Client
    opts := mqtt.NewClientOptions().AddBroker(brokerURL).SetClientID(CLIENT_ID)
    opts.SetKeepAlive(10 * time.Second)
    opts.SetAutoReconnect(true)
    opts.SetOnConnectHandler(func(c mqtt.Client) {
        log.Println("[MQTT] Connected")
    })
    opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
        log.Printf("[MQTT] Connection Lost: %v", err)
    })

    client := mqtt.NewClient(opts)

    // 3. Connect with Retry Logic
    if err := connectWithRetry(client); err != nil {
        log.Fatalf("[Fatal] Exiting: %v", err)
    }

    // 4. Subscribe to "Start Session" Commands
    // This handles the logic: UI says "Start" -> We tell Devices to "Go"
    if token := client.Subscribe(TOPIC_SESSION_START, 1, handleSessionStart); token.Wait() && token.Error() != nil {
        log.Fatalf("[Error] Subscribe failed: %v", token.Error())
    }
    log.Printf("[Subscribed] Listening for commands on: %s", TOPIC_SESSION_START)

    // 5. (Optional) Subscribe to Device Heartbeats to track availability
    // This allows the orchestrator to know who is online before commanding them
    client.Subscribe(TOPIC_DEVICE_STATUS, 1, handleDeviceStatus)
    
    // 6. Block until Signal
    waitForSignal(client)
}

// --- Handlers ---

// handleSessionStart is the Core Orchestration Logic
func handleSessionStart(client mqtt.Client, msg mqtt.Message) {
    var payload SessionStartPayload
    if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
        log.Printf("[Error] Invalid JSON payload: %v", err)
        return
    }

    log.Printf("[Orchestrator] Received Start Request for Session: %s (Group: %s)", 
        payload.SessionID, payload.TargetGroup)

    // A. Lookup Devices in the Target Group
    registry.RLock()
    devices, exists := registry.Groups[payload.TargetGroup]
    registry.RUnlock()

    if !exists || len(devices) == 0 {
        log.Printf("[Warn] No devices found for group '%s'. Aborting.", payload.TargetGroup)
        return
    }

    // B. Create the Command Packet
    cmd := DeviceCommand{
        Action:    "start_test",
        SessionID: payload.SessionID,
        Plan:      payload.TestPlan,
        Config:    "http://firmware-server/config/v2.json", // Example
    }
    
    cmdBytes, _ := json.Marshal(cmd)

    // C. Dispatch Command to Each Device
    for _, deviceID := range devices {
        topic := fmt.Sprintf(TOPIC_DEVICE_CMD, deviceID)
        
        // Publish asynchronously
        go func(t string, d string) {
            token := client.Publish(t, 1, false, cmdBytes)
            token.Wait()
            log.Printf("   -> Sent 'start_test' to Device: %s", d)
        }(topic, deviceID)
    }
}

// handleDeviceStatus updates our knowledge of who is online
func handleDeviceStatus(client mqtt.Client, msg mqtt.Message) {
    // Topic: devices/device_id/status
    // Payload: {"state": "idle", "battery": 98}
    // TODO: Update a real database or Redis here
    // log.Printf("[Heartbeat] %s is online", msg.Topic())
}

// --- Utilities ---

func connectWithRetry(client mqtt.Client) error {
    for i := 0; i < CONNECT_RETRIES; i++ {
        token := client.Connect()
        token.Wait()
        if token.Error() == nil {
            return nil
        }
        log.Printf("[Retry] Connection failed (%d/%d): %v", i+1, CONNECT_RETRIES, token.Error())
        time.Sleep(RETRY_DELAY)
    }
    return fmt.Errorf("timeout connecting to broker")
}

func waitForSignal(client mqtt.Client) {
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
    <-sig
    log.Println("[Shutdown] Disconnecting...")
    client.Disconnect(250)
}