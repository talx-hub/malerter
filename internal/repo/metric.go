package repo

import (
	"strconv"
	"strings"
)

type Metric struct {
	Type   string
	Name   string
	IValue int
	FValue float64
}

func (m *Metric) String() string {
	var value string
	if m.Type == "gauge" {
		value = strconv.FormatFloat(m.FValue, 'f', 2, 64)
	} else {
		value = strconv.Itoa(m.IValue)
	}

	var mSlice = []string{m.Type, m.Name, value}
	return strings.Join(mSlice, "/")
}
