//go:build enterprise
// +build enterprise

package module

import (
	"getsturdy.com/api/pkg/di"
	"getsturdy.com/api/pkg/licenses/enterprise/selfhosted"
)

func Module(c *di.Container) {
	c.Import(selfhosted.Module)
}
