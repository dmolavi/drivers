package dli

import (
	"encoding/json"
	"fmt"

	"github.com/reef-pi/hal"
	"github.com/reef-pi/rpi/i2c"
)

type (
	DLIStrip struct {
		meta     hal.Metadata
		children []*Outlet
		command  *cmd
	}
)

func NewDLIStrip(addr string) *DLIStrip {
	return &DLIStrip{
		meta: hal.Metadata{
			Name:        "dli-pro",
			Description: "Digital Loggers smart power strip driver",
			Capabilities: []hal.Capability{
				hal.DigitalOutput, hal.AnalogInput,
			},
		},
		command: &cmd{
			addr: addr,
		},
		children: make([]*Outlet, 8),
	}
}

func DLIHALAdapter(c []byte, _ i2c.Bus) (hal.Driver, error) {
	var conf Config
	if err := json.Unmarshal(c, &conf); err != nil {
		return nil, err
	}
	s := NewDLIStrip(conf.Address)
	return s, s.FetchSysInfo()
}

func (s *DLIStrip) Metadata() hal.Metadata {
	return s.meta
}

func (s *DLIStrip) Name() string {
	return s.meta.Name
}

func (s *DLIStrip) DigitalOutputPins() []hal.DigitalOutputPin {
	var pins []hal.DigitalOutputPin
	for _, o := range s.children {
		pins = append(pins, o)
	}
	return pins
}

func (s *DLIStrip) DigitalOutputPin(i int) (hal.DigitalOutputPin, error) {
	if i < 0 || i > 5 {
		return nil, fmt.Errorf("invalid pin: %d", i)
	}
	return s.children[i], nil
}

func (s *DLIStrip) Close() error {
	return nil
}
func (s *DLIStrip) FetchSysInfo() error {
	buf, err := s.command.Execute(new(Plug), true)
	if err != nil {
		return err
	}
	var d Plug
	if err := json.Unmarshal(buf, &d); err != nil {
		fmt.Println(string(buf))
		return err
	}
	var children []*Outlet
	for i, ch := range d.System.Sysinfo.Children {
		o := &Outlet{
			name:    ch.Alias,
			id:      ch.ID,
			command: s.command,
			number:  i,
		}
		children = append(children, o)
	}
	s.children = children
	return nil
}

func (s *DLIStrip) Children() []*Outlet {
	return s.children
}

func (p *DLIStrip) AnalogInputPins() []hal.AnalogInputPin {
	var channels []hal.AnalogInputPin
	for _, o := range p.children {
		channels = append(channels, o)
	}
	return channels
}

func (p *DLIStrip) AnalogInputPin(i int) (hal.AnalogInputPin, error) {
	if i < 0 || i > 7 {
		return nil, fmt.Errorf("invalid channel number: %d", i)
	}
	return p.children[i], nil
}

func (p *DLIStrip) Pins(cap hal.Capability) ([]hal.Pin, error) {
	switch cap {
	case hal.DigitalOutput, hal.AnalogInput:
		var channels []hal.Pin
		for _, o := range p.children {
			channels = append(channels, o)
		}
		return channels, nil
	default:
		return nil, fmt.Errorf("unsupported capability:%s", cap.String())
	}
}
