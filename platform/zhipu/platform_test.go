package zhipu

import (
	"fmt"
	"testing"
)

func TestJwt(t *testing.T) {
	result := jwtEncode("47ec79c4702646e2211d8b2c4ce934db.8BxXcm3cgDSSeVWo")
	fmt.Println("jwtEncode", result)
}
