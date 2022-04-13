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
	"golang.org/x/net/context"
)

type Container interface {
	StopAndRemove() error
	GetIP() (string, error)
}

type dockerContainer struct {
	cli dockerClient
	ctx context.Context
	id  string
}

func newContainer(cli dockerClient, ctx context.Context, id string) Container {
	return &dockerContainer{cli: cli, ctx: ctx, id: id}
}

func (c dockerContainer) StopAndRemove() error {
	err := c.cli.ContainerStop(c.ctx, c.id, nil)
	if err != nil {
		return err
	}

	if err := c.cli.ContainerRemove(c.ctx, c.id, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}

func (c dockerContainer) GetIP() (string, error) {
	myContainer, err := c.cli.ContainerInspect(c.ctx, c.id)
	if err != nil {
		return "", err
	}

	if myContainer.NetworkSettings != nil {
		return myContainer.NetworkSettings.IPAddress, nil
	}

	return "", nil
}