package service

import (
	"os"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
)

type program struct {
	stop chan struct{}
}

func (p *program) Start(s service.Service) error {
	p.stop = make(chan struct{})
	go p.run()
	return nil
}

func (p *program) run() {
	logrus.Info("Agent service started")
	StartWorker(p.stop)
}

func (p *program) Stop(s service.Service) error {
	logrus.Info("Agent service stopping")
	close(p.stop)
	return nil
}

func RunService() {
	s, err := newService()
	if err != nil {
		logrus.Fatal(err)
	}

	if len(os.Args) > 1 {
		err = service.Control(s, os.Args[1])
		if err != nil {
			logrus.Fatal(err)
		}
		return
	}

	err = s.Run()
	if err != nil {
		logrus.Fatal(err)
	}
}

func ControlService(action string) error {
	s, err := newService()
	if err != nil {
		return err
	}
	return service.Control(s, action)
}

func GetServiceStatus() (string, error) {
	s, err := newService()
	if err != nil {
		return "unknown", err
	}

	status, err := s.Status()
	if err != nil {
		return "unknown", err
	}

	switch status {
	case service.StatusRunning:
		return "running", nil
	case service.StatusStopped:
		return "stopped", nil
	default:
		return "unknown", nil
	}
}

func newService() (service.Service, error) {
	svcConfig := &service.Config{
		Name:        "MonitorAgent",
		DisplayName: "Monitor Agent Service",
		Description: "System monitoring agent",
	}

	prg := &program{}
	return service.New(prg, svcConfig)
}
