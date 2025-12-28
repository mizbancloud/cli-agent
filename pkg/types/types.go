package types

import (
	"encoding/json"
	"strconv"
)

// NumericBool handles boolean values that come as 0/1 or true/false from API
type NumericBool bool

func (b *NumericBool) UnmarshalJSON(data []byte) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch v := raw.(type) {
	case bool:
		*b = NumericBool(v)
	case float64:
		*b = NumericBool(v != 0)
	case int:
		*b = NumericBool(v != 0)
	case string:
		parsed, _ := strconv.ParseBool(v)
		*b = NumericBool(parsed)
	default:
		*b = false
	}
	return nil
}

func (b NumericBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(b))
}

func (b NumericBool) Bool() bool {
	return bool(b)
}

// FlexibleString handles fields that can come as string or array of strings
// It stores the first value if an array is provided
type FlexibleString string

func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	// Try as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*f = FlexibleString(str)
		return nil
	}

	// Try as array of strings
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		if len(arr) > 0 {
			*f = FlexibleString(arr[0])
		} else {
			*f = ""
		}
		return nil
	}

	// Default to empty
	*f = ""
	return nil
}

func (f FlexibleString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

func (f FlexibleString) String() string {
	return string(f)
}

// NullableInt handles nullable integer fields
type NullableInt struct {
	Value int64
	Valid bool
}

func (n *NullableInt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Value)
}

func (n NullableInt) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

// NullableString handles nullable string fields
type NullableString struct {
	Value string
	Valid bool
}

func (n *NullableString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Value)
}

func (n NullableString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}
