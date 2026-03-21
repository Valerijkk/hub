package tms

// ---------------------------------------------------------------------------
// Типы данных Allure JSON
// ---------------------------------------------------------------------------
// Структуры для десериализации allure-results JSON.
// Все поля используют указатели или omitempty для защитного парсинга —
// ни одно поле не является обязательным, чтобы парсер не падал на
// неполных данных из разных CI-пайплайнов.
// ---------------------------------------------------------------------------

// AllureResult представляет один результат теста в формате Allure JSON.
type AllureResult struct {
	UUID          string               `json:"uuid,omitempty"`
	HistoryID     string               `json:"historyId,omitempty"`
	TestCaseID    string               `json:"testCaseId,omitempty"`
	FullName      string               `json:"fullName,omitempty"`
	Name          string               `json:"name,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetails *AllureStatusDetails  `json:"statusDetails,omitempty"`
	Stage         string               `json:"stage,omitempty"`
	Start         *int64               `json:"start,omitempty"`
	Stop          *int64               `json:"stop,omitempty"`
	Steps         []AllureStep          `json:"steps,omitempty"`
	Labels        []AllureLabel         `json:"labels,omitempty"`
	Links         []AllureLink          `json:"links,omitempty"`
	Parameters    []AllureParameter     `json:"parameters,omitempty"`
	Attachments   []AllureAttachment    `json:"attachments,omitempty"`
	Description   string               `json:"description,omitempty"`
}

// AllureStatusDetails содержит детали статуса (сообщение об ошибке, стектрейс).
type AllureStatusDetails struct {
	Known   bool   `json:"known,omitempty"`
	Message string `json:"message,omitempty"`
	Trace   string `json:"trace,omitempty"`
}

// AllureStep представляет шаг теста (рекурсивная структура).
type AllureStep struct {
	Name        string             `json:"name,omitempty"`
	Status      string             `json:"status,omitempty"`
	Start       *int64             `json:"start,omitempty"`
	Stop        *int64             `json:"stop,omitempty"`
	Steps       []AllureStep       `json:"steps,omitempty"`
	Parameters  []AllureParameter  `json:"parameters,omitempty"`
	Attachments []AllureAttachment `json:"attachments,omitempty"`
}

// AllureLabel представляет метку результата (suite, severity, owner, AS_ID, tms и т.д.).
type AllureLabel struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// AllureLink представляет ссылку (tms, issue, custom).
type AllureLink struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
	Type string `json:"type,omitempty"`
}

// AllureParameter представляет параметр теста.
type AllureParameter struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// AllureAttachment представляет вложение (скриншот, лог и т.д.).
type AllureAttachment struct {
	Name   string `json:"name,omitempty"`
	Source string `json:"source,omitempty"`
	Type   string `json:"type,omitempty"`
}
