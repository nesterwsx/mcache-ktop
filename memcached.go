package main

import (
	"bytes"
	"strconv"
	"strings"
//	"go/types"
//	"fmt"
)

const (
	ERR_NONE = iota
	ERR_NONE_SKIP
	ERR_NO_CMD
	ERR_INVALID_CMD
	ERR_INCOMPLETE_CMD
)

// cmdKeyNoData processes a "get", "incr", "decr", "delete", or "touch" command, all
// of which only allow for a single key to be passed and have no value field.
//
// On the wire, "get" looks like:
//
//     get key\r\n
//
// And "incr", "decr", "touch" look like:
//
//     cmd key value [noreply]\r\n
//
// Where "noreply" is an optional field that indicates whether the server
// should return a response.
func cmdKeyNoData(first_line string) (keys []KeyStat, cmd_err int) {

	var key KeyStat

	split_data := strings.Split(first_line, " ")
	if len(split_data) <= 1 {
		return []KeyStat{}, ERR_INCOMPLETE_CMD
	}

	key.Name = split_data[1]
	if key.Name  == "" {
		return []KeyStat{}, ERR_INCOMPLETE_CMD
	}

	key.RCount = 1
	key.WBytes = uint64(len(first_line) + 2)

	return []KeyStat{key}, ERR_NONE
}

// cmdMultiKeyNoData processes a "gets" command, which allows for
// multiple keys and has no value field.
//
// On the wire, "gets" looks like:
//
//     gets key1 key2 key3\r\n
func cmdMultiKeyNoData(first_line string) (keys []KeyStat, cmd_err int) {

	//var keys []KeyStat

	split_data := strings.Split(first_line, " ")
	if len(split_data) <= 1 {
		return []KeyStat{}, ERR_INCOMPLETE_CMD
	}

	for ndx, key_name := range split_data {
		if ndx == 0 {
			continue // Skip "cmd"
		}
		wb := uint64(len(first_line) + 2)
		keys = append(keys, KeyStat{Name:key_name, RCount:1, WBytes:wb})
	}

	// Return parsed data
	return keys, ERR_NONE
}

// cmdKeyWithData processes a "set", "add", "replace", "append", "cas" or
// "prepend" command, all of which only allow for a single key and have a
// corresponding value field.
//
// These commands look like:
//
//     cmd key flags exptime bytes [noreply]\r\n
//     <data block of `bytes` length>\r\n
//
// Where "noreply" is an optional field that indicates whether the server
// should return a response.
func cmdKeyWithData(first_line string) (keys []KeyStat, cmd_err int) {

	var key KeyStat
	var bytes_str string
	var bytes_ int64
	var err error

	split_data := strings.Split(first_line, " ")
	if len(split_data) != 5 && len(split_data) != 6 {
		return []KeyStat{}, ERR_INCOMPLETE_CMD
	}

	key.Name, bytes_str = split_data[1], split_data[4]

	base := 10
	bitSize := 32
	bytes_, err = strconv.ParseInt(bytes_str, base, bitSize)
	if err != nil {
		return []KeyStat{}, ERR_INVALID_CMD
	}

	key.WCount = 1
	key.WBytes = uint64(bytes_) + uint64(len(first_line) + 4)

	return []KeyStat{key}, ERR_NONE

}

// cmdAnswerValue process responses to "get", "gets" commands
// Answer looks like:
// 	VALUE <key> <flags> <bytes> [<cas unique>]\r\n
// 	<data block>\r\n
//	END\r\n
func cmdAnswerValue(first_line string) (keys []KeyStat, cmd_err int) {

	var bytes_ int64
	var err error
	var key KeyStat
	var bytes_str string

	split_data := strings.Split(first_line, " ")
	if len(split_data) != 4 && len(split_data) != 5 {
		return []KeyStat{}, ERR_INCOMPLETE_CMD
	}

	key.Name, bytes_str = split_data[1], split_data[3]

	base := 10
	bitSize := 32
	bytes_, err = strconv.ParseInt(bytes_str, base, bitSize)
	if err != nil {
		return []KeyStat{}, ERR_INVALID_CMD
	}

	key.RBytes = uint64(bytes_) + uint64(len(first_line) + 4) // What about 'END\r\n'?

	return []KeyStat{key}, ERR_NONE

}

// We don't track to which commands these answers belong
// So we just ignore them
func cmdAnswerIgnore (first_line string) (keys []KeyStat, cmd_err int) {
	return []KeyStat{}, ERR_NONE_SKIP
}


var CMD_PROCESSORS = map[string]func(first_line string) (keys []KeyStat, cmd_err int){
	"get":		cmdKeyNoData,
	"delete":	cmdKeyNoData,
	"gets":		cmdMultiKeyNoData,
	"set":		cmdKeyWithData,
	"cas":		cmdKeyWithData,
	"add":		cmdKeyWithData,
	"replace":	cmdKeyWithData,
	"append":	cmdKeyWithData,
	"prepend":	cmdKeyWithData,
	"incr":		cmdKeyWithData,
	"decr":		cmdKeyWithData,
	"value":	cmdAnswerValue,
	"deleted":	cmdAnswerIgnore,
	"stored":	cmdAnswerIgnore,
	"not_stored":	cmdAnswerIgnore,
	"exists":	cmdAnswerIgnore,
	"not_found":	cmdAnswerIgnore,
	"end":		cmdAnswerIgnore,
}


func parse(app_data []byte) (keys []KeyStat, cmd_err int) {

	space_ndx := bytes.IndexByte(app_data, byte(' '))
	if space_ndx == -1 {
		return []KeyStat{}, ERR_NONE_SKIP
	}

	newline_ndx := bytes.Index(app_data, []byte("\r\n"))
	if newline_ndx == -1 {
		return []KeyStat{}, ERR_NONE_SKIP
	}

	first_line := string(app_data[:newline_ndx])
	split_data := strings.Split(first_line, " ")
	cmd := strings.ToLower(split_data[0])
	if fn, ok := CMD_PROCESSORS[cmd]; ok {
		keys, cmd_err = fn(first_line)
	} else {
		return []KeyStat{}, ERR_INVALID_CMD
	}

	return keys, cmd_err
}
