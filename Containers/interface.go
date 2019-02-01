//Containers Interface
// (For this project we will only use docker but this interface would allow other implementations to be used)
package Containers

type Container struct {
	ID       string            `json:"id"`
	Image    string            `json:"image"`
	Name     string            `json:"name"`
	Labels   map[string]string `json:"labels"`
	Status   string            `json:"status"`
	Env      []string          `json:"env"`
	Ports    map[string]string `json:"ports"`
	Storage  []StorageMount    `json:"storage"`
	Networks []Network         `json:"networks"`
	Alias    string            `json:"alias"`
	Stats    Stats             `json:"stats"`
}

type Network struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

type Stats struct {
	CpuUsage float64 `json:"cpu_usage"`
}

type StorageMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mount_path"`
}

//TODO Add ports to Interface

type ContainerRuntime interface {
	AssertOnline() error
	CreateContainer(container *Container) error
	ReadContainer(id string) (*Container, error)
	ReadAllContainers() ([]*Container, error)
	ReadAllContainersWithFilter(filter map[string]map[string]bool) ([]*Container, error)
	UpdateContainer(container *Container) error
	DeleteContainer(id string) error
	DeleteContainerWithFilter(filter map[string]map[string]bool) error
	CreateStorage(name string) error
	DeleteStorage(name string) error
	CreateNetwork(name string) error
	DeleteNetwork(name string) error
	NetworkExists(name string) (bool, error)
}
