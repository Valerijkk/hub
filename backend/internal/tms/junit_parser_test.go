package tms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const junitTestSuitesXML = `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="LoginSuite" tests="3" failures="1" errors="0" skipped="1">
    <testcase classname="com.example.LoginTest" name="testSuccess" time="0.120"/>
    <testcase classname="com.example.LoginTest" name="testFail" time="0.050">
      <failure message="assertion failed" type="AssertionError">expected true</failure>
    </testcase>
    <testcase classname="com.example.LoginTest" name="testSkip" time="0.010">
      <skipped message="not implemented"/>
    </testcase>
  </testsuite>
</testsuites>`

const junitSingleSuiteXML = `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="SingleSuite" tests="1">
  <testcase classname="com.example.BrokenTest" name="testBroken" time="0.200">
    <error message="runtime error" type="RuntimeError">panic: boom</error>
  </testcase>
</testsuite>`

func TestParseJUnitXML_TestSuites(t *testing.T) {
	results, err := ParseJUnitXML([]byte(junitTestSuitesXML))
	require.NoError(t, err)
	require.Len(t, results, 3)

	assert.Equal(t, "testSuccess", results[0].Name)
	assert.Equal(t, "passed", results[0].Status)
	assert.Equal(t, "com.example.LoginTest.testSuccess", results[0].FullName)
	assert.Equal(t, "LoginSuite", extractLabelValue(results[0].Labels, "suite"))

	assert.Equal(t, "failed", results[1].Status)
	require.NotNil(t, results[1].StatusDetails)
	assert.Equal(t, "assertion failed", results[1].StatusDetails.Message)

	assert.Equal(t, "skipped", results[2].Status)
}

func TestParseJUnitXML_SingleSuite(t *testing.T) {
	results, err := ParseJUnitXML([]byte(junitSingleSuiteXML))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "broken", results[0].Status)
	require.NotNil(t, results[0].StatusDetails)
	assert.Contains(t, results[0].StatusDetails.Trace, "panic")
}

func TestParseJUnitXML_EmptySuite(t *testing.T) {
	_, err := ParseJUnitXML([]byte(`<testsuites></testsuites>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "testsuite")
}

func TestParseJUnitXML_Malformed(t *testing.T) {
	_, err := ParseJUnitXML([]byte(`<not xml`))
	require.Error(t, err)
}

func TestJUnitTestCaseToAllure(t *testing.T) {
	tests := []struct {
		name           string
		suite          string
		tc             JUnitTestCase
		wantStatus     string
		wantFullName   string
		wantMessage    string
		wantSuiteLabel string
	}{
		{
			name:  "passed",
			suite: "AuthSuite",
			tc: JUnitTestCase{
				Name:      "testOk",
				ClassName: "com.example.AuthTest",
				Time:      0.5,
			},
			wantStatus:     "passed",
			wantFullName:   "com.example.AuthTest.testOk",
			wantSuiteLabel: "AuthSuite",
		},
		{
			name:  "failed with message",
			suite: "AuthSuite",
			tc: JUnitTestCase{
				Name:      "testFail",
				ClassName: "com.example.AuthTest",
				Failure:   &JUnitFailure{Message: "expected true"},
			},
			wantStatus:     "failed",
			wantFullName:   "com.example.AuthTest.testFail",
			wantSuiteLabel: "AuthSuite",
			wantMessage:    "expected true",
		},
		{
			name:  "skipped",
			suite: "AuthSuite",
			tc: JUnitTestCase{
				Name:    "testSkip",
				Skipped: &JUnitSkipped{Message: "wip"},
			},
			wantStatus:     "skipped",
			wantFullName:   "testSkip",
			wantSuiteLabel: "AuthSuite",
			wantMessage:    "wip",
		},
		{
			name:  "name only",
			suite: "Minimal",
			tc: JUnitTestCase{
				Name: "standalone",
			},
			wantStatus:   "passed",
			wantFullName: "standalone",
			wantSuiteLabel: "Minimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := junitTestCaseToAllure(tt.suite, tt.tc)
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantFullName, got.FullName)
			assert.Equal(t, tt.wantSuiteLabel, extractLabelValue(got.Labels, "suite"))
			if tt.wantMessage != "" {
				require.NotNil(t, got.StatusDetails)
				assert.Equal(t, tt.wantMessage, got.StatusDetails.Message)
			}
		})
	}
}

func TestParseJUnitXML_EmptyTestCases(t *testing.T) {
	_, err := ParseJUnitXML([]byte(`<testsuites><testsuite name="Empty" tests="0"/></testsuites>`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "testcase")
}
