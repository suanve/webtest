package engine

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

//创建一个对象
func CreateClient(host string) *client.Client {
	//tcp://127.0.0.1:2376
	cli, err := client.NewClient(host, "v1.12", nil, nil)
	// defer cli.Close()
	if err != nil {
		fmt.Println("连接docker失败")
	}
	return cli
}

// 列出镜像
func ListImage(cli *client.Client) {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	Log(err)

	for _, image := range images {
		fmt.Println(image)
	}
}

// 列出容器
func ListContainer(cli *client.Client) []types.Container {
	images, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	Log(err)
	// for _, image := range images {
	// 	fmt.Println(image)
	// }
	return images
}

// 创建容器
func CreateContainer(cli *client.Client, imageName string, iutPort int, outPort int, userName string, challengeId int) string {
	exports := make(nat.PortSet, 10)
	port, err := nat.NewPort("tcp", strconv.Itoa(iutPort))
	Log(err)
	exports[port] = struct{}{}
	config := &container.Config{Image: string(imageName), ExposedPorts: exports}

	portBind := nat.PortBinding{HostPort: strconv.Itoa(outPort)}
	portMap := make(nat.PortMap, 0)
	tmp := make([]nat.PortBinding, 0, 1)
	tmp = append(tmp, portBind)
	portMap[port] = tmp
	hostConfig := &container.HostConfig{PortBindings: portMap}

	containerName := userName + "_" + strconv.Itoa(challengeId)
	body, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, containerName)
	Log(err)
	fmt.Printf("容器ID: %s\n", body.ID)
	return body.ID
}

// 启动
func StartContainer(containerID string, cli *client.Client) {
	err := cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	Log(err)
	if err == nil {
		fmt.Println("容器", containerID, "启动成功")
	}
}

// 停止
func StopContainer(containerID string, cli *client.Client) {
	timeout := time.Second * 10
	fmt.Println(containerID)
	err := cli.ContainerStop(context.Background(), containerID, &timeout)
	if err != nil {
		Log(err)
	} else {
		fmt.Printf("容器%s已经被停止\n", containerID)
	}
}

// 删除
func RemoveContainer(containerID string, cli *client.Client) (string, error) {
	err := cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
	Log(err)
	return containerID, err
}

func Log(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}
}
