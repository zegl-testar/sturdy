package cloud

import (
	"getsturdy.com/api/pkg/di"
	"getsturdy.com/api/pkg/users/enterprise/cloud/graphql"
	"getsturdy.com/api/pkg/users/enterprise/cloud/service"
)

func Module(c *di.Container) {
	c.Import(service.Module)
	c.Import(graphql.Module)
}
