//go:build cloud
// +build cloud

package module

import (
	"getsturdy.com/api/pkg/configuration/enterprise/cloud"
	"getsturdy.com/api/pkg/di"
)

func Module(c *di.Container) {
	c.Register(cloud.New)
}
