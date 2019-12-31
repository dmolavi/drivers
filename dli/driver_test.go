package dli

import (
	"testing"
)

func TestDLIWebProSwitch(t *testing.T) {
	p := NewDLIWebProSwitch("pro.digital-loggers.com:5002","admin","4321")
	p.digestGet("http://","pro.digital-loggers.com:5002", "/restapi/relay/outlets/")
}
