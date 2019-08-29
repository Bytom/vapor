package api

import (
	"github.com/gin-gonic/gin"

	"github.com/vapor/toolbar/precog/database/orm"
	serverCommon "github.com/vapor/toolbar/server"
)

type listNodesReq struct{ serverCommon.Display }

func (s *Server) ListNodes(c *gin.Context, listNodesReq *listNodesReq, query *serverCommon.PaginationQuery) ([]*orm.Node, error) {
	return nil, nil
}