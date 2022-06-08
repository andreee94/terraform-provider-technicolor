package provider

import (
	"fmt"
	"strings"
)

func computeID(name string, wanPort int, protocol string) string {
	return fmt.Sprintf("%s/%s/%d", strings.ToLower(name), protocol, wanPort)
}
