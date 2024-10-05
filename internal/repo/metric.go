package repo

import (
	"strconv"
	"strings"
)

type Metric struct {
	Type  string
	Name  string
	Value any
}

func (m *Metric) String() string {
	var value string
	if m.Type == "gauge" {
		value = strconv.FormatFloat(m.Value.(float64), 'f', 2, 64)
	} else {
		value = strconv.Itoa(m.Value.(int))
	}

	var mSlice = []string{m.Type, m.Name, value}
	return strings.Join(mSlice, "/")
}

func (m *Metric) Update(other Metric) {
	if m.Type == "gauge" {
		m.Value = other.Value
	} else {
		updated := m.Value.(int) + other.Value.(int)
		m.Value = updated
	}
}
