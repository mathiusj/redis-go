package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parser parses RESP protocol messages
type Parser struct {
	reader *bufio.Reader
}

// NewParser creates a new RESP parser
func NewParser(r io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(r),
	}
}

// Parse reads and parses the next RESP value
func (p *Parser) Parse() (Value, error) {
	b, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch Type(b) {
	case SimpleString:
		return p.parseSimpleString()
	case Error:
		return p.parseError()
	case Integer:
		return p.parseInteger()
	case BulkString:
		return p.parseBulkString()
	case Array:
		return p.parseArray()
	default:
		return Value{}, fmt.Errorf("unknown RESP type: %c", b)
	}
}

func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Remove \r\n
	return strings.TrimSuffix(line, "\r\n"), nil
}

func (p *Parser) parseSimpleString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: SimpleString, Str: line}, nil
}

func (p *Parser) parseError() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: Error, Str: line}, nil
}

func (p *Parser) parseInteger() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}

	i, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid integer: %s", line)
	}

	return Value{Type: Integer, Integer: i}, nil
}

func (p *Parser) parseBulkString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid bulk string length: %s", line)
	}

	if length == -1 {
		// Null bulk string - use special marker
		return Value{Type: BulkString, Str: "\x00NULL"}, nil
	}

	if length < 0 {
		return Value{}, fmt.Errorf("invalid bulk string length: %d", length)
	}

	// Read the bulk string data
	data := make([]byte, length+2) // +2 for \r\n
	_, err = io.ReadFull(p.reader, data)
	if err != nil {
		return Value{}, err
	}

	// Remove \r\n
	return Value{Type: BulkString, Str: string(data[:length])}, nil
}

func (p *Parser) parseArray() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}

	count, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid array count: %s", line)
	}

	if count == -1 {
		// Null array
		return Value{Type: Array, Array: []Value{}}, nil
	}

	if count < 0 {
		return Value{}, fmt.Errorf("invalid array count: %d", count)
	}

	array := make([]Value, count)
	for i := 0; i < count; i++ {
		value, err := p.Parse()
		if err != nil {
			return Value{}, err
		}
		array[i] = value
	}

	return Value{Type: Array, Array: array}, nil
}
