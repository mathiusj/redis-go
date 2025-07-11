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
	IsNull  bool    // Indicates if this is a null value (for bulk strings or arrays)
}

// String returns a string representation of the value
func (value Value) String() string {
	switch value.Type {
	case SimpleString, Error:
		return value.Str
	case BulkString:
		if value.IsNull {
			return ""
		}
		return value.Str
	case Integer:
		return fmt.Sprintf("%d", value.Integer)
	case Array:
		return fmt.Sprintf("%v", value.Array)
	default:
		return ""
	}
}

// IsError returns true if the value is an error
func (value Value) IsError() bool {
	return value.Type == Error
}

// GetCommand extracts the command name from an array value
func (value Value) GetCommand() (string, error) {
	if value.Type != Array || len(value.Array) == 0 {
		return "", fmt.Errorf("invalid command format")
	}

	cmd := value.Array[0]
	if cmd.Type != BulkString {
		return "", fmt.Errorf("command must be a bulk string")
	}

	return cmd.Str, nil
}

// GetArgs returns the arguments from an array value (excluding the command)
func (value Value) GetArgs() []string {
	if value.Type != Array || len(value.Array) <= 1 {
		return []string{}
	}

	args := make([]string, 0, len(value.Array)-1)
	for index := 1; index < len(value.Array); index++ {
		args = append(args, value.Array[index].String())
	}

	return args
}
