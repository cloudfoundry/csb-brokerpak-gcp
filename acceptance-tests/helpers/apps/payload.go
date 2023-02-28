package apps

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/gomega"
)

type Payload string

func (p Payload) Parse(receiver any) {
	Expect(reflect.ValueOf(receiver).Kind()).To(Equal(reflect.Ptr), "must pass a pointer to the receiver")
	Expect(json.Unmarshal([]byte(p), receiver)).To(Succeed())
}
