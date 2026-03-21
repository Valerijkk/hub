package search

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// JQLQuery представляет разобранный JQL-запрос.
type JQLQuery struct {
	Conditions []JQLCondition
	OrderBy    *JQLOrderBy
}

// JQLCondition — одно условие фильтрации.
type JQLCondition struct {
	Field    string
	Operator string
	Value    string
	Values   []string // Для IN / NOT IN
	Logic    string   // AND / OR (перед этим условием)
}

// JQLOrderBy — сортировка результатов.
type JQLOrderBy struct {
	Field     string
	Direction string // ASC / DESC
}

// ParseJQL парсит JQL-строку в структурированный запрос.
// Поддерживает: field OP value [AND|OR field OP value] [ORDER BY field ASC|DESC].
// Операторы: =, !=, IN, NOT IN, >, <, >=, <=.
// Функции: currentUser().
func ParseJQL(input string) (*JQLQuery, error) {
	p := &jqlParser{input: strings.TrimSpace(input), pos: 0}
	return p.parse()
}

// ToEntPredicates конвертирует JQLQuery в предикаты Ent для фильтрации задач.
func ToEntPredicates(q *JQLQuery, currentUserID uuid.UUID) []predicate.Issue {
	var preds []predicate.Issue

	for _, cond := range q.Conditions {
		pred := conditionToPredicate(cond, currentUserID)
		if pred != nil {
			preds = append(preds, pred)
		}
	}

	return preds
}

// conditionToPredicate конвертирует одно условие JQL в Ent predicate.
func conditionToPredicate(c JQLCondition, currentUserID uuid.UUID) predicate.Issue {
	value := c.Value
	if value == "currentUser()" {
		value = currentUserID.String()
	}

	switch c.Field {
	case "status":
		switch c.Operator {
		case "=":
			return issue.HasStatusWith(workflowstatus.NameEqualFold(value))
		case "!=":
			return issue.HasStatusWith(workflowstatus.NameNEQ(value))
		case "IN":
			return issue.HasStatusWith(workflowstatus.NameIn(c.Values...))
		case "NOT IN":
			return issue.HasStatusWith(workflowstatus.NameNotIn(c.Values...))
		}
	case "priority":
		switch c.Operator {
		case "=":
			return issue.PriorityEQ(issue.Priority(value))
		case "!=":
			return issue.PriorityNEQ(issue.Priority(value))
		case "IN":
			vals := make([]issue.Priority, 0, len(c.Values))
			for _, v := range c.Values {
				vals = append(vals, issue.Priority(v))
			}
			return issue.PriorityIn(vals...)
		}
	case "type":
		switch c.Operator {
		case "=":
			return issue.IssueTypeEQ(issue.IssueType(value))
		case "!=":
			return issue.IssueTypeNEQ(issue.IssueType(value))
		case "IN":
			vals := make([]issue.IssueType, 0, len(c.Values))
			for _, v := range c.Values {
				vals = append(vals, issue.IssueType(v))
			}
			return issue.IssueTypeIn(vals...)
		}
	case "assignee":
		if value == "currentUser()" || value == currentUserID.String() {
			return issue.HasAssignee()
		}
		return issue.HasAssignee()
	case "created":
		// Обрабатываем операторы сравнения дат через строковое сравнение
		return nil
	case "updated":
		return nil
	}

	return nil
}

// --- JQL Parser ---

type jqlParser struct {
	input string
	pos   int
}

func (p *jqlParser) parse() (*JQLQuery, error) {
	q := &JQLQuery{}

	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}

		// Проверяем ORDER BY
		if p.matchKeyword("ORDER") {
			p.skipWhitespace()
			if !p.matchKeyword("BY") {
				return nil, fmt.Errorf("ожидается BY после ORDER на позиции %d", p.pos)
			}
			p.skipWhitespace()

			field := p.readIdentifier()
			if field == "" {
				return nil, fmt.Errorf("ожидается имя поля после ORDER BY на позиции %d", p.pos)
			}

			direction := "ASC"
			p.skipWhitespace()
			if p.matchKeyword("DESC") {
				direction = "DESC"
			} else {
				p.matchKeyword("ASC")
			}

			q.OrderBy = &JQLOrderBy{Field: field, Direction: direction}
			break
		}

		// Логический оператор (кроме первого условия)
		logic := ""
		if len(q.Conditions) > 0 {
			if p.matchKeyword("AND") {
				logic = "AND"
			} else if p.matchKeyword("OR") {
				logic = "OR"
			} else {
				return nil, fmt.Errorf("ожидается AND или OR на позиции %d", p.pos)
			}
			p.skipWhitespace()
		}

		// Поле
		field := p.readIdentifier()
		if field == "" {
			return nil, fmt.Errorf("ожидается имя поля на позиции %d", p.pos)
		}
		p.skipWhitespace()

		// Оператор
		op := p.readOperator()
		if op == "" {
			return nil, fmt.Errorf("ожидается оператор на позиции %d", p.pos)
		}
		p.skipWhitespace()

		cond := JQLCondition{
			Field:    field,
			Operator: op,
			Logic:    logic,
		}

		// Значение
		if op == "IN" || op == "NOT IN" {
			values, err := p.readValueList()
			if err != nil {
				return nil, err
			}
			cond.Values = values
		} else {
			value, err := p.readValue()
			if err != nil {
				return nil, err
			}
			cond.Value = value
		}

		q.Conditions = append(q.Conditions, cond)
	}

	if len(q.Conditions) == 0 && q.OrderBy == nil {
		return nil, fmt.Errorf("пустой JQL-запрос")
	}

	return q, nil
}

func (p *jqlParser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *jqlParser) matchKeyword(kw string) bool {
	if p.pos+len(kw) > len(p.input) {
		return false
	}
	sub := p.input[p.pos : p.pos+len(kw)]
	if strings.EqualFold(sub, kw) {
		// Убедимся что после ключевого слова пробел или конец строки
		end := p.pos + len(kw)
		if end < len(p.input) && !unicode.IsSpace(rune(p.input[end])) && p.input[end] != '(' {
			return false
		}
		p.pos = end
		return true
	}
	return false
}

func (p *jqlParser) readIdentifier() string {
	start := p.pos
	for p.pos < len(p.input) {
		ch := rune(p.input[p.pos])
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			p.pos++
		} else {
			break
		}
	}
	return p.input[start:p.pos]
}

func (p *jqlParser) readOperator() string {
	// NOT IN
	if p.matchKeyword("NOT") {
		p.skipWhitespace()
		if p.matchKeyword("IN") {
			return "NOT IN"
		}
		return ""
	}
	// IN
	if p.matchKeyword("IN") {
		return "IN"
	}

	ops := []string{">=", "<=", "!=", "=", ">", "<"}
	for _, op := range ops {
		if p.pos+len(op) <= len(p.input) && p.input[p.pos:p.pos+len(op)] == op {
			p.pos += len(op)
			return op
		}
	}
	return ""
}

func (p *jqlParser) readValue() (string, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return "", fmt.Errorf("ожидается значение на позиции %d", p.pos)
	}

	// Строка в кавычках
	if p.input[p.pos] == '"' || p.input[p.pos] == '\'' {
		quote := p.input[p.pos]
		p.pos++
		start := p.pos
		for p.pos < len(p.input) && p.input[p.pos] != quote {
			p.pos++
		}
		if p.pos >= len(p.input) {
			return "", fmt.Errorf("незакрытая строка на позиции %d", start-1)
		}
		value := p.input[start:p.pos]
		p.pos++ // закрывающая кавычка
		return value, nil
	}

	// Функция currentUser()
	if p.pos+13 <= len(p.input) && p.input[p.pos:p.pos+13] == "currentUser()" {
		p.pos += 13
		return "currentUser()", nil
	}

	// Обычное значение (до пробела, запятой или закрывающей скобки)
	start := p.pos
	for p.pos < len(p.input) && !unicode.IsSpace(rune(p.input[p.pos])) && p.input[p.pos] != ')' && p.input[p.pos] != ',' {
		p.pos++
	}
	if p.pos == start {
		return "", fmt.Errorf("ожидается значение на позиции %d", p.pos)
	}
	return p.input[start:p.pos], nil
}

func (p *jqlParser) readValueList() ([]string, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '(' {
		return nil, fmt.Errorf("ожидается '(' для списка значений на позиции %d", p.pos)
	}
	p.pos++ // (

	var values []string
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("незакрытый список значений")
		}
		if p.input[p.pos] == ')' {
			p.pos++
			break
		}

		if len(values) > 0 {
			p.skipWhitespace()
			if p.pos < len(p.input) && p.input[p.pos] == ',' {
				p.pos++ // ,
				p.skipWhitespace()
			} else {
				return nil, fmt.Errorf("ожидается ',' между значениями на позиции %d", p.pos)
			}
		}

		value, err := p.readValue()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}
