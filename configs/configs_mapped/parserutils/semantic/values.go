/*
ORM mapper for Freelancer ini reader. Easy mapping values to change.
*/
package semantic

import (
	"encoding/json"

	"github.com/darklab8/fl-configs/configs/configs_mapped/parserutils/inireader"
	"github.com/darklab8/fl-configs/configs/configs_settings/logus"
	"github.com/darklab8/go-typelog/typelog"
)

// ORM values

type ValueType int64

const (
	TypeComment ValueType = iota
	TypeVisible
)

type Value struct {
	section    *inireader.Section
	key        string
	optional   bool
	value_type ValueType
	order      int
	index      int
}

func NewValue(
	section *inireader.Section,
	key string,
) *Value {
	return &Value{
		section:    section,
		key:        key,
		value_type: TypeVisible,
	}
}

func (v Value) isComment() bool {
	return v.value_type == TypeComment
}

type ValueOption func(i *Value)

func Order(order int) ValueOption {
	return func(i *Value) {
		i.order = order
	}
}

func Index(index int) ValueOption {
	return func(i *Value) {
		i.index = index
	}
}

func Optional() ValueOption {
	return func(i *Value) {
		i.optional = true
	}
}

func Comment() ValueOption {
	return func(i *Value) {
		i.value_type = TypeComment
	}
}

func quickJson(value any) string {
	result, err := json.Marshal(value)
	if err != nil {
		return err.Error()
	}
	return string(result)
}

func handleGetCrashReporting(value *Value) {
	if r := recover(); r != nil {
		if value == nil {
			logus.Log.Panic("value is not defined. not possible. ;)")
			return
		} else {
			logus.Log.Error("unable to Get() from semantic.",
				typelog.Any("value", quickJson(value)),
				typelog.Any("key", value.key),
				typelog.NestedStruct("section", value.section),
			)
		}
		panic(r)
	}
}
