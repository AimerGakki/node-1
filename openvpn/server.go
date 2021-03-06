package openvpn

import "sync"

func NewServer(config *ServerConfig, directoryRuntime string, middlewares ...ManagementMiddleware) *Server {
	// Add the management interface socketAddress to the config
	socketAddress := tempFilename(directoryRuntime, "openvpn-management-", ".sock")
	config.SetManagementSocket(socketAddress)

	return &Server{
		config:     config,
		management: NewManagement(socketAddress, "[server-management] ", middlewares...),
		process:    NewProcess("[server-openvpn] "),
	}
}

type Server struct {
	config     *ServerConfig
	management *Management
	process    *Process
}

func (server *Server) Start() error {
	// Start the management interface (if it isnt already started)
	if err := server.management.Start(); err != nil {
		return err
	}

	// Fetch the current params
	arguments, err := ConfigToArguments(*server.config.Config)
	if err != nil {
		return err
	}

	return server.process.Start(arguments)
}

func (client *Server) Wait() error {
	return client.process.Wait()
}

func (server *Server) Stop() {
	waiter := sync.WaitGroup{}

	waiter.Add(1)
	go func() {
		defer waiter.Done()
		server.process.Stop()
	}()

	waiter.Add(1)
	go func() {
		defer waiter.Done()
		server.management.Stop()
	}()

	waiter.Wait()
}
