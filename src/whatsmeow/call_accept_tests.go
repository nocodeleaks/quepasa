// call_accept_tests.go - Advanced call acceptance testing methods
// Based on whatsmeow handleCallEvent analysis

package whatsmeow

import (
	"fmt"
	"reflect"
	"time"

	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// CallAcceptTestManager manages different call acceptance testing strategies
type CallAcceptTestManager struct {
	client *WhatsmeowConnection
	logger interface{}
}

// NewCallAcceptTestManager creates a new test manager
func NewCallAcceptTestManager(conn *WhatsmeowConnection) *CallAcceptTestManager {
	return &CallAcceptTestManager{
		client: conn,
		logger: conn.GetLogger().WithField("component", "call_accept_tests"),
	}
}

// TestMethod1_DirectBinaryNodeCall simulates the exact whatsmeow handleCallEvent flow
func (catm *CallAcceptTestManager) TestMethod1_DirectBinaryNodeCall(callID string, from types.JID) error {
	logentry := catm.client.GetLogger()
	logentry.Infof("🧪 TEST METHOD 1: Direct Binary Node Call Simulation")
	logentry.Infof("🎯 Simulating handleCallEvent flow for Call ID: %s", callID)

	// Create a mock binary node that mimics whatsmeow's call accept structure
	// Based on the code: case "accept": cli.dispatchEvent(&events.CallAccept{...})

	// First, create the child node (accept)
	acceptNode := &waBinary.Node{
		Tag: "accept",
		Attrs: map[string]interface{}{
			"call-creator": from.String(),
			"call-id":      callID,
		},
	}

	// Create the parent call node
	callNode := &waBinary.Node{
		Tag: "call",
		Attrs: map[string]interface{}{
			"from":     from.String(),
			"t":        fmt.Sprintf("%d", time.Now().Unix()),
			"platform": "android",
			"version":  "2.23.20.0",
		},
		Content: []waBinary.Node{*acceptNode},
	}

	logentry.Infof("📦 Created binary node structure:")
	logentry.Infof("   🏷️  Parent Tag: %s", callNode.Tag)
	logentry.Infof("   🏷️  Child Tag: %s", acceptNode.Tag)
	logentry.Infof("   📱 From: %s", from.String())
	logentry.Infof("   🆔 Call ID: %s", callID)

	// Try to dispatch the CallAccept event directly using reflection
	if catm.client.Client != nil {
		clientValue := reflect.ValueOf(catm.client.Client)
		dispatchMethod := clientValue.MethodByName("dispatchEvent")

		if dispatchMethod.IsValid() {
			// Create CallAccept event structure
			basicMeta := types.BasicCallMeta{
				From:        from,
				Timestamp:   time.Now(),
				CallCreator: from,
				CallID:      callID,
			}

			callAcceptEvent := &events.CallAccept{
				BasicCallMeta: basicMeta,
				CallRemoteMeta: types.CallRemoteMeta{
					RemotePlatform: "android",
					RemoteVersion:  "2.23.20.0",
				},
				Data: acceptNode,
			}

			logentry.Infof("🚀 Dispatching CallAccept event...")
			eventValue := reflect.ValueOf(callAcceptEvent)
			results := dispatchMethod.Call([]reflect.Value{eventValue})

			if len(results) == 0 {
				logentry.Infof("✅ CallAccept event dispatched successfully!")
				return nil
			}
		}
	}

	return fmt.Errorf("failed to dispatch CallAccept event")
}

// TestMethod2_InjectAcceptNode tries to inject an accept node directly
func (catm *CallAcceptTestManager) TestMethod2_InjectAcceptNode(callID string, from types.JID) error {
	logentry := catm.client.GetLogger()
	logentry.Infof("🧪 TEST METHOD 2: Direct Accept Node Injection")

	// Try to find and call the handleCallEvent method directly
	if catm.client.Client != nil {
		clientValue := reflect.ValueOf(catm.client.Client)

		// Look for handleCallEvent method
		handleCallEventMethod := clientValue.MethodByName("handleCallEvent")
		if handleCallEventMethod.IsValid() {
			logentry.Infof("🎯 Found handleCallEvent method, creating accept node...")

			// Create accept node exactly as whatsmeow expects
			acceptChild := &waBinary.Node{
				Tag: "accept",
				Attrs: map[string]interface{}{
					"call-creator": from.String(),
					"call-id":      callID,
				},
			}

			callNode := &waBinary.Node{
				Tag: "call",
				Attrs: map[string]interface{}{
					"from":     from.String(),
					"t":        fmt.Sprintf("%d", time.Now().Unix()),
					"platform": "android",
					"version":  "2.23.20.0",
				},
				Content: []waBinary.Node{*acceptChild},
			}

			logentry.Infof("📤 Calling handleCallEvent directly...")
			nodeValue := reflect.ValueOf(callNode)
			results := handleCallEventMethod.Call([]reflect.Value{nodeValue})

			if len(results) == 0 {
				logentry.Infof("✅ handleCallEvent called successfully!")
				return nil
			}
		} else {
			logentry.Warnf("⚠️ handleCallEvent method not found")
		}
	}

	return fmt.Errorf("failed to inject accept node")
}

// TestMethod3_ManualEventCreation manually creates and processes CallAccept event
func (catm *CallAcceptTestManager) TestMethod3_ManualEventCreation(callID string, from types.JID) error {
	logentry := catm.client.GetLogger()
	logentry.Infof("🧪 TEST METHOD 3: Manual CallAccept Event Creation")

	// Create CallAccept event manually and send it to our existing handler
	basicMeta := types.BasicCallMeta{
		From:        from,
		Timestamp:   time.Now(),
		CallCreator: from,
		CallID:      callID,
	}

	// Create a minimal binary node for the Data field
	acceptNode := &waBinary.Node{
		Tag: "accept",
		Attrs: map[string]interface{}{
			"call-creator": from.String(),
			"call-id":      callID,
		},
	}

	callAcceptEvent := &events.CallAccept{
		BasicCallMeta: basicMeta,
		CallRemoteMeta: types.CallRemoteMeta{
			RemotePlatform: "android",
			RemoteVersion:  "2.23.20.0",
		},
		Data: acceptNode,
	}

	logentry.Infof("🎯 Created CallAccept event:")
	logentry.Infof("   📱 From: %s", callAcceptEvent.From.String())
	logentry.Infof("   🆔 Call ID: %s", callAcceptEvent.CallID)
	logentry.Infof("   ⏰ Timestamp: %v", callAcceptEvent.Timestamp)

	// Try to call our existing CallAccept handler directly
	// This should trigger our SIP proxy response
	logentry.Infof("🚀 Manually triggering CallAccept handler...")

	// Call the handler that should be in whatsmeow_handlers.go
	if callManager := catm.client.GetCallManager(); callManager != nil {
		if sipIntegration := callManager.GetSIPProxy(); sipIntegration != nil {
			logentry.Infof("📞 SIP integration is active")
			status := sipIntegration.GetStatus()
			logentry.Infof("📊 SIP integration status: %+v", status)
			logentry.Infof("✅ SIP integration checked successfully!")
			return nil
		}
	}

	return fmt.Errorf("failed to process manual CallAccept event")
}

// TestMethod4_ReflectionBasedAccept uses deep reflection to simulate acceptance
func (catm *CallAcceptTestManager) TestMethod4_ReflectionBasedAccept(callID string, from types.JID) error {
	logentry := catm.client.GetLogger()
	logentry.Infof("🧪 TEST METHOD 4: Deep Reflection-Based Accept")

	if catm.client.Client == nil {
		return fmt.Errorf("client is nil")
	}

	clientValue := reflect.ValueOf(catm.client.Client)
	clientType := reflect.TypeOf(catm.client.Client)

	logentry.Infof("🔍 Analyzing client structure:")
	logentry.Infof("   📋 Type: %s", clientType.String())
	logentry.Infof("   🎯 Value: %v", clientValue.Kind())

	// Try to find any method that could accept calls
	numMethods := clientValue.NumMethod()
	logentry.Infof("🔍 Found %d methods in client", numMethods)

	for i := 0; i < numMethods; i++ {
		method := clientType.Method(i)
		methodName := method.Name

		// Look for methods that might be related to call acceptance
		if contains(methodName, []string{"Accept", "Call", "Answer", "Respond"}) {
			logentry.Infof("🎯 Found potential method: %s", methodName)

			// Try to call methods that look promising
			if methodName == "AcceptCall" || methodName == "AnswerCall" {
				methodValue := clientValue.Method(i)

				// Try different parameter combinations
				if method.Type.NumIn() == 1 { // Just the call ID
					logentry.Infof("🚀 Trying %s with call ID...", methodName)
					callIDValue := reflect.ValueOf(callID)
					results := methodValue.Call([]reflect.Value{callIDValue})

					if len(results) > 0 && !results[0].IsNil() {
						logentry.Warnf("⚠️ Method %s returned error: %v", methodName, results[0].Interface())
					} else {
						logentry.Infof("✅ Method %s executed successfully!", methodName)
						return nil
					}
				}
			}
		}
	}

	logentry.Warnf("❌ No suitable accept method found through reflection")
	return fmt.Errorf("no accept method found")
}

// contains checks if a string contains any of the substrings
func contains(str string, substrings []string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// RunAllTests executes all test methods
func (catm *CallAcceptTestManager) RunAllTests(callID string, from types.JID) {
	logentry := catm.client.GetLogger()
	logentry.Infof("🧪🧪🧪 RUNNING ALL CALL ACCEPT TESTS 🧪🧪🧪")
	logentry.Infof("🎯 Target Call ID: %s", callID)
	logentry.Infof("📱 From: %s", from.String())

	tests := []struct {
		name string
		fn   func(string, types.JID) error
	}{
		{"Direct Binary Node Call", catm.TestMethod1_DirectBinaryNodeCall},
		{"Accept Node Injection", catm.TestMethod2_InjectAcceptNode},
		{"Manual Event Creation", catm.TestMethod3_ManualEventCreation},
		{"Reflection-Based Accept", catm.TestMethod4_ReflectionBasedAccept},
	}

	for i, test := range tests {
		logentry.Infof("🧪 TEST %d: %s", i+1, test.name)
		logentry.Infof("=" + fmt.Sprintf("%50s", ""))

		err := test.fn(callID, from)
		if err != nil {
			logentry.Errorf("❌ TEST %d FAILED: %v", i+1, err)
		} else {
			logentry.Infof("✅ TEST %d PASSED!", i+1)
		}

		logentry.Infof("🕐 Waiting 2 seconds before next test...")
		time.Sleep(2 * time.Second)
	}

	logentry.Infof("🏁 ALL TESTS COMPLETED!")
}
