package tms

import (
	"encoding/xml"
	"fmt"
)

// ---------------------------------------------------------------------------
// Парсер JUnit XML
// ---------------------------------------------------------------------------
// Поддерживает стандартный формат JUnit XML (testsuites > testsuite > testcase).
// Преобразует результаты в формат AllureResult для единого пайплайна обработки.
// ---------------------------------------------------------------------------

// JUnitTestSuites — корневой элемент JUnit XML (<testsuites>).
type JUnitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	TestSuites []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite — набор тестов (<testsuite>).
type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Name       string          `xml:"name,attr"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Errors     int             `xml:"errors,attr"`
	Skipped    int             `xml:"skipped,attr"`
	Time       float64         `xml:"time,attr"`
	Timestamp  string          `xml:"timestamp,attr"`
	TestCases  []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase — один тест-кейс в JUnit XML (<testcase>).
type JUnitTestCase struct {
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure"`
	Error     *JUnitError   `xml:"error"`
	Skipped   *JUnitSkipped `xml:"skipped"`
}

// JUnitFailure — элемент failure (ожидаемая ошибка).
type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

// JUnitError — элемент error (неожиданная ошибка).
type JUnitError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

// JUnitSkipped — элемент skipped.
type JUnitSkipped struct {
	Message string `xml:"message,attr"`
}

// ParseJUnitXML парсит JUnit XML и преобразует в AllureResult для единого пайплайна.
func ParseJUnitXML(data []byte) ([]AllureResult, error) {
	// Пробуем как <testsuites>
	var suites JUnitTestSuites
	if err := xml.Unmarshal(data, &suites); err != nil {
		// Пробуем как одиночный <testsuite>
		var single JUnitTestSuite
		if err2 := xml.Unmarshal(data, &single); err2 != nil {
			return nil, fmt.Errorf("ошибка парсинга JUnit XML: %w (также пробовали testsuite: %v)", err, err2)
		}
		suites.TestSuites = []JUnitTestSuite{single}
	}

	// Если обёртка <testsuites> пустая, но внутри были testsuite
	if len(suites.TestSuites) == 0 {
		return nil, fmt.Errorf("JUnit XML не содержит ни одного testsuite")
	}

	var results []AllureResult
	for _, suite := range suites.TestSuites {
		for _, tc := range suite.TestCases {
			result := junitTestCaseToAllure(suite.Name, tc)
			results = append(results, result)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("JUnit XML не содержит ни одного testcase")
	}

	return results, nil
}

// junitTestCaseToAllure конвертирует JUnit testcase в AllureResult.
func junitTestCaseToAllure(suiteName string, tc JUnitTestCase) AllureResult {
	status := "passed"
	var statusDetails *AllureStatusDetails

	switch {
	case tc.Error != nil:
		status = "broken"
		statusDetails = &AllureStatusDetails{
			Message: tc.Error.Message,
			Trace:   tc.Error.Body,
		}
	case tc.Failure != nil:
		status = "failed"
		statusDetails = &AllureStatusDetails{
			Message: tc.Failure.Message,
			Trace:   tc.Failure.Body,
		}
	case tc.Skipped != nil:
		status = "skipped"
		if tc.Skipped.Message != "" {
			statusDetails = &AllureStatusDetails{
				Message: tc.Skipped.Message,
			}
		}
	}

	// Длительность в миллисекундах
	durationMs := int64(tc.Time * 1000)
	var start, stop *int64
	if durationMs > 0 {
		s := int64(0)
		e := durationMs
		start = &s
		stop = &e
	}

	// fullName = className.methodName
	fullName := tc.ClassName
	if tc.Name != "" {
		if fullName != "" {
			fullName += "." + tc.Name
		} else {
			fullName = tc.Name
		}
	}

	labels := []AllureLabel{
		{Name: "suite", Value: suiteName},
	}
	if tc.ClassName != "" {
		labels = append(labels, AllureLabel{Name: "testClass", Value: tc.ClassName})
	}

	return AllureResult{
		Name:          tc.Name,
		FullName:      fullName,
		Status:        status,
		StatusDetails: statusDetails,
		Start:         start,
		Stop:          stop,
		Labels:        labels,
	}
}
