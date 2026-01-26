package orgmcp

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

func (u *Uid) String() string {
	return u.uid
}

func (u *Uid) MarshalText() ([]byte, error) {
	return []byte(u.uid), nil
}
