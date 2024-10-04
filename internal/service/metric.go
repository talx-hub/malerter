package service

import (
	"strconv"
	"strings"
)

type Metric struct {
	mType  string
	name   string
	iValue int
	fValue float64
}

func (m *Metric) String() string {
	var value string
	if m.mType == "gauge" {
		value = strconv.FormatFloat(m.fValue, 'f', 2, 64)
	} else {
		value = strconv.Itoa(m.iValue)
	}

	var mSlice = []string{m.mType, m.name, value}
	return strings.Join(mSlice, "/")
}
