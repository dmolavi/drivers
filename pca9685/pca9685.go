package pca9685

import (
	"math"
	"time"

	"github.com/reef-pi/rpi/i2c"
)

const (
	clockFreq        = 25000000
	pwmControlPoints = 4096
	mode1RegAddr     = 0x00
	mode2RegAddr     = 0x01
	pwm0OnLowReg     = 0x6
	defaultFreq      = 490
	OUTDRV           = 0x0
	SLEEP            = byte(0x10)
	ALLCALL          = 0x01
	ALL_LED_ON_L     = 0xFA
	ALL_LED_ON_H     = 0xFB
	ALL_LED_OFF_L    = 0xFC
	ALL_LED_OFF_H    = 0xFD
	preScaleRegAddr  = byte(0xFE)
	RESTART          = 0x80
)

type PCA9685 struct {
	addr byte
	bus  i2c.Bus
	Freq int
}

func New(addr byte, bus i2c.Bus) *PCA9685 {
	return &PCA9685{
		addr: addr,
		bus:  bus,
		Freq: defaultFreq,
	}
}

func (p *PCA9685) Setup() error {
	if err := p.SetAll(0, 0); err != nil {
		return err
	}
	if err := p.bus.WriteToReg(p.addr, mode2RegAddr, []byte{OUTDRV}); err != nil {
		return nil
	}
	if err := p.bus.WriteToReg(p.addr, mode1RegAddr, []byte{ALLCALL}); err != nil {
		return nil
	}
	time.Sleep(time.Millisecond)
	mode1 := make([]byte, 1)
	if err := p.bus.ReadFromReg(p.addr, mode1RegAddr, mode1); err != nil {
		return err
	}
	newMode := mode1[0] & ^SLEEP
	if err := p.bus.WriteToReg(p.addr, mode1RegAddr, []byte{newMode}); err != nil {
		return err
	}
	time.Sleep(time.Millisecond)
	return nil
}

func (p *PCA9685) SetFrequency() error {
	if p.Freq == 0 {
		p.Freq = defaultFreq
	}
	preScaleValue := int(math.Floor(float64(clockFreq/(pwmControlPoints*p.Freq))+float64(0.5)) - 1)

	oldMode := make([]byte, 1)
	if err := p.bus.ReadFromReg(p.addr, mode1RegAddr, oldMode); err != nil {
		return err
	}
	newMode := (oldMode[0] & 0x7F) | SLEEP
	if err := p.bus.WriteToReg(p.addr, mode1RegAddr, []byte{newMode}); err != nil {
		return err
	}
	time.Sleep(time.Millisecond)
	if err := p.bus.WriteToReg(p.addr, preScaleRegAddr, []byte{byte(preScaleValue)}); err != nil {
		return err
	}
	if err := p.bus.WriteToReg(p.addr, mode1RegAddr, oldMode); err != nil {
		return err
	}
	time.Sleep(time.Millisecond)
	if err := p.bus.WriteToReg(p.addr, mode1RegAddr, []byte{oldMode[0] | RESTART}); err != nil {
		return err
	}
	return nil
}

func (p *PCA9685) SetPwm(channel int, on, off uint16) error {
	if off > 4095 {
		off = 4095
	}
	// Split the ints into 4 bytes
	chanReg := byte(pwm0OnLowReg + (4 * channel))
	onLow := byte(on & 0xFF)
	onHigh := byte(on >> 8)
	offLow := byte(off & 0xFF)
	offHigh := byte(off >> 8)

	//log.Println("onLow ", onTimeLow, " onHigh ", onTimeHigh, " offLow ", offTimeLow, " offHigh ", offTimeHigh)
	if err := p.bus.WriteToReg(p.addr, chanReg, []byte{onLow}); err != nil {
		return err
	}
	if err := p.bus.WriteToReg(p.addr, chanReg+1, []byte{onHigh}); err != nil {
		return err
	}
	if err := p.bus.WriteToReg(p.addr, chanReg+2, []byte{offLow}); err != nil {
		return err
	}
	return p.bus.WriteToReg(p.addr, chanReg+3, []byte{offHigh})
}

func (p *PCA9685) Close() error {
	// Clear all channels to full off
	for regAddr := 0x06; regAddr < 0x50; regAddr += 4 {
		if err := p.bus.WriteToReg(p.addr, byte(regAddr), []byte{0x00, 0x00, 0x00, 0x10}); err != nil {
			return err
		}
	}
	return nil
}

func (p *PCA9685) SetAll(on, off uint16) error {
	if err := p.bus.WriteToReg(p.addr, ALL_LED_ON_L, []byte{byte(on & 0xFF)}); err != nil {
		return err
	}
	if err := p.bus.WriteToReg(p.addr, ALL_LED_ON_H, []byte{byte(on >> 8)}); err != nil {
		return err
	}
	if err := p.bus.WriteToReg(p.addr, ALL_LED_OFF_L, []byte{byte(off & 0xFF)}); err != nil {
		return err
	}
	if err := p.bus.WriteToReg(p.addr, ALL_LED_OFF_H, []byte{byte(off >> 8)}); err != nil {
		return err
	}
	return nil
}
