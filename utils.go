package main

import (
  "sort"
  "time"
  "strconv"
  "math/rand"
  "encoding/base64"
)

// URL encodes reserved and non-ASCII characters in a string, following the
// "path" convention whereby spaces are converted to %20.
func PercentEncode(str string) string {
	hex := 0
	for i := range str {
		if shouldEncode(str[i]) {
			hex++
		}
	}

	HEXCHARS := "0123456789ABCDEF"
	buf := make([]byte, len(str) + 2 * hex)
	j := 0
	for i := range str {
		if shouldEncode(str[i]) {
			buf[j] = '%'
			buf[j + 1] = HEXCHARS[str[i] >> 4]
			buf[j + 2] = HEXCHARS[str[i] & 0xF]
			j += 2
		} else {
			buf[j] = str[i]
		}
		j++
	}

	return string(buf)
}

func shouldEncode(c byte) bool {
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}
	switch c {
	case '-', '.', '_', '~':
		return false
	default:
		return true
	}
}

// Take the union of two maps and return a copy
func CombinedMaps(src, dst map[string]string) map[string]string {
  copy := make(map[string]string)
  for k, v := range dst {
    copy[k] = v
  }
  for k, v := range src {
    copy[k] = v
  }
  return copy
}

// Return a list of the sorted keys of a map
func GetSortedKeys(m map[string]string) []string {
  keys := make([]string, len(m))
  i := 0
  for k, _ := range m {
    keys[i] = k
    i++
  }
  sort.Strings(keys)
  return keys
}

// Panic if error exists
func CheckError(e error) {
  if e != nil {
    panic(e)
  }
}

// Get numBytes rando, bytes and base64 encode them into a string
func GetRandomB64(numBytes uint) string {
  raw := make([]byte, numBytes)
  _, e := rand.Read(raw)
  CheckError(e)
  return base64.StdEncoding.EncodeToString(raw)
}

// Get the unix timestamp as a string
func GetUnixTimestamp() string {
  var seconds int64 = time.Now().Unix()
  return strconv.FormatInt(seconds, 10)
}
