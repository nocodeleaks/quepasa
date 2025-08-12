package sipproxy

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// SIPProxyCallAcceptTestManager manages call acceptance testing
type SIPProxyCallAcceptTestManager struct {
	logger      *log.Entry
	proxy       *SIPProxyManager
	testResults []SIPProxyCallTestResult
	isRunning   bool
	maxTests    int
}

// NewSIPProxyCallAcceptTestManager creates a new call acceptance test manager
func NewSIPProxyCallAcceptTestManager(proxy *SIPProxyManager) *SIPProxyCallAcceptTestManager {
	return &SIPProxyCallAcceptTestManager{
		logger:      log.WithField("component", "call_test"),
		proxy:       proxy,
		testResults: make([]SIPProxyCallTestResult, 0),
		isRunning:   false,
		maxTests:    100, // Keep last 100 test results
	}
}

// RunCallTest runs a single call acceptance test
func (catm *SIPProxyCallAcceptTestManager) RunCallTest(testID, fromPhone, toPhone string) *SIPProxyCallTestResult {
	result := SIPProxyCallTestResult{
		TestID:    testID,
		CallID:    fmt.Sprintf("test-%s-%d", testID, time.Now().Unix()),
		StartTime: time.Now(),
		Success:   false,
	}

	catm.logger.Infof("🧪 Running call test %s (CallID: %s)", testID, result.CallID)

	// Send test INVITE
	if err := catm.proxy.SendSIPInvite(fromPhone, toPhone, result.CallID); err != nil {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.ErrorMsg = err.Error()
		catm.logger.Errorf("❌ Call test %s failed: %v", testID, err)
	} else {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.Success = true
		result.SIPResponse = "200 OK"
		catm.logger.Infof("✅ Call test %s completed successfully", testID)
	}

	catm.addTestResult(result)
	return &result
}

// addTestResult adds a test result to the manager
func (catm *SIPProxyCallAcceptTestManager) addTestResult(result SIPProxyCallTestResult) {
	catm.testResults = append(catm.testResults, result)

	// Keep only the last maxTests results
	if len(catm.testResults) > catm.maxTests {
		catm.testResults = catm.testResults[1:]
	}
}

// GetTestResults returns all test results
func (catm *SIPProxyCallAcceptTestManager) GetTestResults() []SIPProxyCallTestResult {
	return catm.testResults
}

// GetSuccessfulTests returns only successful test results
func (catm *SIPProxyCallAcceptTestManager) GetSuccessfulTests() []SIPProxyCallTestResult {
	var successful []SIPProxyCallTestResult
	for _, result := range catm.testResults {
		if result.Success {
			successful = append(successful, result)
		}
	}
	return successful
}

// GetFailedTests returns only failed test results
func (catm *SIPProxyCallAcceptTestManager) GetFailedTests() []SIPProxyCallTestResult {
	var failed []SIPProxyCallTestResult
	for _, result := range catm.testResults {
		if !result.Success {
			failed = append(failed, result)
		}
	}
	return failed
}

// GetTestSuccessRate returns the success rate of tests as a percentage
func (catm *SIPProxyCallAcceptTestManager) GetTestSuccessRate() float64 {
	if len(catm.testResults) == 0 {
		return 0.0
	}

	successful := len(catm.GetSuccessfulTests())
	return float64(successful) / float64(len(catm.testResults)) * 100.0
}

// ClearTestResults clears all test results
func (catm *SIPProxyCallAcceptTestManager) ClearTestResults() {
	catm.testResults = make([]SIPProxyCallTestResult, 0)
	catm.logger.Info("🧹 Test results cleared")
}

// IsRunning returns whether tests are currently running
func (catm *SIPProxyCallAcceptTestManager) IsRunning() bool {
	return catm.isRunning
}

// SetMaxTests sets the maximum number of test results to keep
func (catm *SIPProxyCallAcceptTestManager) SetMaxTests(max int) {
	catm.maxTests = max
}
