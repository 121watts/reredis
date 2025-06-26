package wal

import (
	"strconv"
)

//  - * = Array (your commands)
//  - $ = Bulk String (your keys/values)
//  - + = Simple String (status messages)
//  - : = Integer (numbers)
//  - - = Error (error messages)

//  SET mykey myvalue becomes:
//  *3\r\n          // Array with 3 elements
//  $3\r\nSET\r\n   // "SET" (3 bytes)
//  $5\r\nmykey\r\n // "mykey" (5 bytes)
//  $7\r\nmyvalue\r\n // "myvalue" (7 bytes)

func EncodeArray(strings []string) []byte {
	length := len(strings)
	lengthStr := strconv.Itoa(length)
	var result []byte
	initialLine := []byte("*" + lengthStr + "\r\n")
	result = append(result, initialLine...)

	for _, s := range strings {
		result = append(result, EncodeBulkString(s)...)
	}

	return result
}

func EncodeBulkString(s string) []byte {
	length := len(s)
	lengthStr := strconv.Itoa(length)
	result := []byte("$" + lengthStr + "\r\n" + s + "\r\n")

	return result
}
