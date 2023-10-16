package agent

import "github.com/kardianos/service"

var logger service.Logger

type program struct{}

func (p *program) run() error {
	// work goes here

	return nil
}

func (p *program) Start(s service.Service) error {
	go p.run()

	return nil
}

func (p *program) Stop(s service.Service) error {

	return nil
}

/*func (p *program) Restart() error {
	// TODO implement me
	panic("implement me")
}

func (p *program) Install() error {
	// TODO implement me
	panic("implement me")
}

func (p *program) Uninstall() error {
	// TODO implement me
	panic("implement me")
}
*/
