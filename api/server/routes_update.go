package server

import (
	"context"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/runner/task"
	"github.com/iron-io/runner/common"
)

func (s *Server) handleRouteUpdate(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	log := common.Logger(ctx)

	var wroute models.RouteWrapper

	err := c.BindJSON(&wroute)
	if err != nil {
		log.WithError(err).Debug(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrInvalidJSON))
		return
	}

	if wroute.Route == nil {
		log.Debug(models.ErrRoutesMissingNew)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrRoutesMissingNew))
		return
	}

	if wroute.Route.Path != "" {
		log.Debug(models.ErrRoutesPathImmutable)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrRoutesPathImmutable))
		return
	}

	wroute.Route.AppName = ctx.Value("appName").(string)
	wroute.Route.Path = path.Clean(ctx.Value("routePath").(string))

	if wroute.Route.Image != "" {
		err = s.Runner.EnsureImageExists(ctx, &task.Config{
			Image: wroute.Route.Image,
		})
		if err != nil {
			log.WithError(err).Debug(models.ErrRoutesUpdate)
			c.JSON(http.StatusBadRequest, simpleError(models.ErrUsableImage))
			return
		}
	}

	route, err := s.Datastore.UpdateRoute(ctx, wroute.Route)
	if err == models.ErrRoutesNotFound {
		log.WithError(err).Debug(models.ErrRoutesUpdate)
		c.JSON(http.StatusNotFound, simpleError(err))
		return
	} else if err != nil {
		log.WithError(err).Error(models.ErrRoutesUpdate)
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrRoutesUpdate))
		return
	}

	c.JSON(http.StatusOK, routeResponse{"Route successfully updated", route})
}
