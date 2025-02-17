package cloud

import (
	"getsturdy.com/api/pkg/di"
	"getsturdy.com/api/pkg/installations/enterprise/cloud/graphql"
)

func Module(c *di.Container) {
	c.Import(graphql.Module)
}
