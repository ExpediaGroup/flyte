/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"time"
)

type Docker interface {
	// Run can be called with the following to set environment variables and exposed ports
	// Run("containerName","imagePath",[]string{"MONGO_HOST=172.0.0.1"}, []string{"8080/tcp","127.0.0.1"})
	// This utilises  portSpecTemplate and the expected format for port specifications "ip:hostPort:containerPort"
	Run(name, imagePath string, envVars []string, hostconfig []string) (Container, error)
	Pull(imagePath string) error
}

type dockerClient interface {
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
}

type docker struct {
	cli dockerClient
	ctx context.Context
}

var newDockerClient = func() (dockerClient, error) {
	return client.NewEnvClient()
}

func NewDocker() (Docker, error) {
	cli, err := newDockerClient()
	if err != nil {
		return nil, err
	}

	docker := &docker{
		cli: cli,
		ctx: context.Background(),
	}

	return docker, nil
}

func (d *docker) Run(name, imagePath string, envVars []string, hostconfig []string) (Container, error) {
	hasImage, err := d.hasImage(imagePath)
	if err != nil {
		return nil, err
	}

	if !hasImage {
		if err := d.Pull(imagePath); err != nil {
			return nil, err
		}
	}

	return d.startContainer(imagePath, name, envVars, hostconfig)
}

func (d *docker) hasImage(imagePath string) (bool, error) {
	filter := filters.NewArgs()
	filter.Add("reference", imagePath)

	images, err := d.cli.ImageList(d.ctx, types.ImageListOptions{Filters: filter})
	if err != nil {
		return false, err
	}
	return len(images) > 0, nil
}

func (d *docker) Pull(imagePath string) error {
	return d.imagePull(imagePath)
}

func (d *docker) startContainer(imagePath, name string, envVars []string, hostconfig []string) (Container, error) {

	portSet, portBindings, _ := nat.ParsePortSpecs(hostconfig)

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
	}

	containerConfig := &container.Config{Image: imagePath, Env: envVars, ExposedPorts: portSet}
	resp, err := d.cli.ContainerCreate(d.ctx, containerConfig, hostConfig, nil, name)
	if err != nil {
		return nil, err
	}

	if err = d.cli.ContainerStart(d.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	return newContainer(d.cli, d.ctx, resp.ID), err
}

func (d *docker) imagePull(imagePath string) error {
	out, err := d.cli.ImagePull(d.ctx, imagePath, types.ImagePullOptions{})

	if err != nil {
		log.Err(err).Msg("unable to pull image")
		if out != nil {
			out.Close()
		}
		return err
	}
	defer out.Close()

	log.Info().Msgf("Pulling and reading image, may take several minutes: %s", imagePath)
	_, err = io.Copy(ioutil.Discard, out)
	if err != nil {
		log.Err(err).Msg("unable to pull image")
		return err
	}

	return nil
}