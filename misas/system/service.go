package system

type Service any

type Services map[string]Service

func WithServices(s Services) Option {
	return func(system *System) {
		for n, serv := range s {
			WithService(n, serv)(system)
			system.Services[n] = serv
		}
	}
}

func WithService(name string, s Service) Option {
	return func(system *System) {
		system.Services[name] = s
	}
}
