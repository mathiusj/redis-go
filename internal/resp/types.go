package resp

import "fmt"

// Type represents the type of RESP value
type Type byte

const (
	SimpleString Type = '+'
	Error        Type = '-'
	Integer      Type = ':'
	BulkString   Type = '$'
	Array        Type = '*'
)

// Value represents a RESP value
type Value struct {
	Type    Type
	Str     string  // Renamed from String to avoid conflict with String() method
	Integer int
	Array   []Value
}

// String returns a string representation of the value
func (v Value) String() string {
	switch v.Type {
	case SimpleString, Error, BulkString:
		return v.Str
	case Integer:
		return fmt.Sprintf("%d", v.Integer)
	case Array:
		return fmt.Sprintf("%v", v.Array)
	default:
		return ""
	}
}

// IsError returns true if the value is an error
func (v Value) IsError() bool {
	return v.Type == Error
}

// GetCommand extracts the command name from an array value
func (v Value) GetCommand() (string, error) {
	if v.Type != Array || len(v.Array) == 0 {
		return "", fmt.Errorf("invalid command format")
	}

	cmd := v.Array[0]
	if cmd.Type != BulkString {
		return "", fmt.Errorf("command must be a bulk string")
	}

	return cmd.Str, nil
}

// GetArgs returns the arguments from an array value (excluding the command)
func (v Value) GetArgs() []string {
	if v.Type != Array || len(v.Array) <= 1 {
		return []string{}
	}

	args := make([]string, 0, len(v.Array)-1)
	for i := 1; i < len(v.Array); i++ {
		args = append(args, v.Array[i].String())
	}

	return args
}
