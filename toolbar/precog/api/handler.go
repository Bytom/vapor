package api

import (
	"github.com/gin-gonic/gin"

	"github.com/vapor/toolbar/precog/database/orm"
	serverCommon "github.com/vapor/toolbar/server"
)

type listNodesReq struct{ serverCommon.Display }

// TODO: fix PaginationQuery???
func (s *Server) ListNodes(c *gin.Context, listNodesReq *listNodesReq, query *serverCommon.PaginationQuery) ([]*orm.Node, error) {
	var ormNodes []*orm.Node
	if err := s.db.Offset(query.Start).Limit(query.Limit).Find(&ormNodes).Error; err != nil {
		return nil, err
	}

	return ormNodes, nil
}
