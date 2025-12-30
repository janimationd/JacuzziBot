package models

import (
	"encoding/json"
)

// A user and all information we track that's associated with them.
// There's other info we use that we explicitly *don't* track here, such as display names and pronouns,
// since those can change and it would be bad if we cached the wrong thing.
type User struct {
	// For Discord, this is a value like "404740897331478539"
	UserId string
	// Points are a float behind the scenes, but we should limit precision
	// when showing to users.
	Points float64
}

func (u User) ToJsonBytes() ([]byte, error) {
	return json.Marshal(u)
}

func (u User) ToJsonString() (string, error) {
	bytes, err := u.ToJsonBytes()
	str := string(bytes)
	return str, err
}

func FromJsonBytes(bytes []byte) (User, error) {
	var u User
	err := json.Unmarshal(bytes, &u)
	return u, err
}

func FromJsonString(str string) (User, error) {
	return FromJsonBytes([]byte(str))
}
