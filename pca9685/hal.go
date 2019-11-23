package pca9685

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/reef-pi/hal"
	"github.com/reef-pi/rpi/i2c"
)

const (
	_name        = "pca9685"
	_description = "HAL Driver for PCA9685 16 channel PWM IC"
)

type PCA9685Config struct {
	Address   int `json:"address"` // 0x40
	Frequency int `json:"frequency"`
}

type pca9685Channel struct {
	driver  *pca9685Driver
	channel int
	v       float64
}

func (c *pca9685Channel) Name() string { return fmt.Sprintf("%d", c.channel) }
func (c *pca9685Channel) Number() int  { return c.channel }
func (c *pca9685Channel) Close() error { return nil }
func (c *pca9685Channel) Set(value float64) error {
	if err := c.driver.set(c.channel, value); err != nil {
		return err
	}
	c.v = value
	return nil
}
func (c *pca9685Channel) Write(b bool) error {
	var v float64
	if b {
		v = 100
	}
	if err := c.driver.set(c.channel, v); err != nil {
		return err
	}
	c.v = v
	return nil
}

func (c *pca9685Channel) LastState() bool { return c.v == 100 }

type pca9685Driver struct {
	config   PCA9685Config
	hwDriver *PCA9685
	mu       *sync.Mutex
	channels []*pca9685Channel
}

var DefaultPCA9685Config = PCA9685Config{
	Address:   0x40,
	Frequency: 1500,
}

func HALAdapter(c []byte, bus i2c.Bus) (hal.Driver, error) {
	config := DefaultPCA9685Config
	if err := json.Unmarshal(c, &config); err != nil {
		return nil, err
	}

	hwDriver := New(byte(config.Address), bus)
	pwm := pca9685Driver{
		config:   config,
		mu:       &sync.Mutex{},
		hwDriver: hwDriver,
	}
	if config.Frequency == 0 {
		log.Println("WARNING: pca9685 driver pwm frequency set to 0. Falling back to 1500")
		config.Frequency = 1500
	}
	hwDriver.Freq = config.Frequency // overriding default

	// Create the 16 channels the hardware has
	for i := 0; i < 16; i++ {
		ch := &pca9685Channel{
			channel: i,
			driver:  &pwm,
		}
		pwm.channels = append(pwm.channels, ch)
	}

	return &pwm, hwDriver.Setup()
}

func (p *pca9685Driver) Close() error {
	return p.hwDriver.Close()
}

func (p *pca9685Driver) Metadata() hal.Metadata {
	return hal.Metadata{
		Name:        _name,
		Description: _description,
		Capabilities: []hal.Capability{
			hal.PWM, hal.DigitalOutput,
		},
	}
}

func (p *pca9685Driver) PWMChannels() []hal.PWMChannel {
	// Return array of channels soreted by name
	var chs []hal.PWMChannel
	for _, ch := range p.channels {
		chs = append(chs, ch)
	}
	sort.Slice(chs, func(i, j int) bool { return chs[i].Name() < chs[j].Name() })
	return chs
}

func (p *pca9685Driver) PWMChannel(chnum int) (hal.PWMChannel, error) {
	// Return given channel
	if chnum < 0 || chnum >= len(p.channels) {
		return nil, fmt.Errorf("invalid channel %d", chnum)
	}
	return p.channels[chnum], nil
}
func (p *pca9685Driver) DigitalOutputPins() []hal.DigitalOutputPin {
	pins := make([]hal.DigitalOutputPin, len(p.channels))
	for i, ch := range p.channels {
		pins[i] = ch
	}
	return pins
}

func (p *pca9685Driver) DigitalOutputPin(n int) (hal.DigitalOutputPin, error) {
	return p.PWMChannel(n)
}

// value should be within 0-100
func (p *pca9685Driver) set(pin int, value float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch {
	case value > 100:
		return fmt.Errorf("invalid pwm range: %f, value should be less than 100", value)
	case value < 0:
		return fmt.Errorf("invalid pwm range: %f, value should be greater than 100", value)
	case value == 0:
		return p.hwDriver.SetPwm(pin, 0, 4096)
	case value == 100:
		return p.hwDriver.SetPwm(pin, 4096, 0)
	default:
		return p.hwDriver.SetPwm(pin, 0, uint16(value*40.95))
	}
}

func (p *pca9685Driver) Pins(cap hal.Capability) ([]hal.Pin, error) {
	switch cap {
	case hal.DigitalOutput, hal.PWM:
		var pins []hal.Pin
		for _, pin := range p.channels {
			pins = append(pins, pin)
		}
		return pins, nil
	default:
		return nil, fmt.Errorf("unsupported capability: %s", cap.String())
	}
}
