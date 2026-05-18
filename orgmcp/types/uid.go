package orgmcp_types

import (
	"fmt"
)

type UidValue interface {
	~int | ~string
}

type Uid struct {
	uid string
}

func NewUid[T UidValue](uid T) Uid {
	return Uid{uid: fmt.Sprintf("%v", uid)}
}

func (u Uid) String() string {
	return u.uid
}

func (u *Uid) MarshalText() ([]byte, error) {
	return []byte(u.uid), nil
}

func (u *Uid) UnmarshalText(text []byte) error {
	u.uid = string(text)
	return nil
}

func (u Uid) GetSchema() map[string]any {
	return map[string]any{"type": "string"}
}
