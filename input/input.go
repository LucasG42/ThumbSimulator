package input

import "os"

var Buffer string

func LoadFile(input string) bool {
	data, err := os.ReadFile(input)
	if err == nil {
		Buffer = string(data)
		return true
	}
	return false
}
