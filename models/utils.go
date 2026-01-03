package models

import "encoding/json"

func ToJsonBytes[T any](obj T) ([]byte, error) {
	return json.Marshal(obj)
}

func ToJsonString[T any](obj T) (string, error) {
	bytes, err := ToJsonBytes(obj)
	str := string(bytes)
	return str, err
}

func FromJsonBytes[T any](bytes []byte) (T, error) {
	var obj T
	err := json.Unmarshal(bytes, &obj)
	return obj, err
}

func FromJsonString[T any](str string) (T, error) {
	return FromJsonBytes[T]([]byte(str))
}
