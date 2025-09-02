package context

import (
	"s29-be/pkg/cache"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ServiceContext struct {
	dbContext      *gorm.DB
	router         *fiber.App
	publicRouter   *fiber.Router
	internalRouter *fiber.Router

	cacheClient *cache.Client
}

func NewServiceContext(dbContext *gorm.DB, router *fiber.App, publicRouter *fiber.Router, internalRouter *fiber.Router, cacheClient *cache.Client) *ServiceContext {
	return &ServiceContext{
		dbContext:      dbContext,
		router:         router,
		publicRouter:   publicRouter,
		internalRouter: internalRouter,
		cacheClient:    cacheClient,
	}
}

func (ctx ServiceContext) GetDB() *gorm.DB {
	return ctx.dbContext
}

func (ctx ServiceContext) GetRouter() *fiber.App {
	return ctx.router
}

func (ctx ServiceContext) GetPublicRouter() *fiber.Router {
	return ctx.publicRouter
}

func (ctx ServiceContext) GetInternalRouter() *fiber.Router {
	return ctx.internalRouter
}

func (ctx ServiceContext) GetCacheClient() *cache.Client {
	return ctx.cacheClient
}
