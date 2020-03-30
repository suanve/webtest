package engine

import (
	"webtest/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var cli *client.Client

func init() {
	cli = CreateClient(config.DockerHost)
}

func Ctr_ListContainer() []Container {
	//CreateClient(config.DockerHost)
	// data, _ := ioutil.ReadAll(c.Request.Body)
	var containers []Container
	var container Container
	// CreateContainer(cli)
	for _, image := range ListContainer(cli) {
		// image.
		// json.Unmarshal([]byte(image), &container)
		container.ID = image.ID
		container.Names = image.Names
		container.Image = image.Image
		container.ImageID = image.ImageID
		container.Command = image.Command
		container.Created = image.Created
		container.PortS = image.Ports
		container.State = image.State
		containers = append(containers, container)
	}
	return containers
}

//获取传入的信息，启动一个容器，返回对应的容器id
func Ctr_CreateContainer(imageName string, iutPort int, outPort int, userName string, challengeId int) string {
	//创建一个容器
	id := CreateContainer(cli, imageName, iutPort, outPort, userName, challengeId)
	//启动它
	StartContainer(id, cli)

	return id
}

func Ctr_StopContainer(containerID string) string {
	//停止一个容器
	StopContainer(containerID, cli)
	//删除它
	id, _ := RemoveContainer(containerID, cli)
	return id
}

type Container struct {
	ID      string       `json:id`
	Names   []string     `json:name`
	Image   string       `json:image`
	ImageID string       `json:imageid`
	Command string       `json:command`
	Created int64        `json:created`
	PortS   []types.Port `json:ports`
	State   string       `json:state`
}
