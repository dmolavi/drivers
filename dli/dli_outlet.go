package dli

import (
	"github.com/reef-pi/hal"
)

type (
	Outlet struct {
		name       string
		id         string
		command    *cmd
		state      bool
		calibrator hal.Calibrator		
		number     int
	}
)

func (o *Outlet) Name() string {
	return o.name
}
func (o *Outlet) Number() int {
	return o.number
}

func (o *Outlet) Write(state bool) error {
	if state {
		return o.On()
	}
	return o.Off()
}

func (o *Outlet) LastState() bool {
	return o.state
}

func (o *Outlet) On() error {
	cmd := new(CmdRelayState)
	cmd.System.RelayState.State = 1
	cmd.Context.Children = []string{o.id}
	if _, err := o.command.Execute(cmd, false); err != nil {
		return err
	}
	o.state = true
	return nil
}
func (o *Outlet) Off() error {
	cmd := new(CmdRelayState)
	cmd.System.RelayState.State = 0
	cmd.Context.Children = []string{o.id}
	if _, err := o.command.Execute(cmd, false); err != nil {
		return err
	}
	o.state = true
	return nil
}

func (o *Outlet) Calibrate(points []hal.Measurement) error {
	return nil
}

func (o *Outlet) Read() (float64, error) {
	return 0,nil
}

func (o *Outlet) Measure() (float64, error) {
	return 0,nil
}

func (o *Outlet) Close() error {
	return nil
}
