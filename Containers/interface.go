//Containers Interface
// (For this project we will only use docker but this interface would allow other implementations to be used)
package Containers

type Container struct {
	ID     string            `json:"id"`
	Image  string            `json:"image"`
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Status string            `json:"status"`
}

//TODO Add ports to Interface

type ContainerRuntime interface {
	CreateContainer(container *Container) error
	ReadContainer(id string) (*Container, error)
	ReadAllContainers() ([]*Container, error)
	UpdateContainer(container *Container) error
	DeleteContainer(id string) error
}
