package host

import "github.com/gin-gonic/gin"

type hostRouter struct{}

func NewRouter(ginEngine *gin.Engine) {
	h := &hostRouter{}
	h.initRoutes(ginEngine)
}

func (h *hostRouter) initRoutes(ginEngine *gin.Engine) {
	hostRoute := ginEngine.Group("/hosts")
	{
		hostRoute.POST("", h.addHost)        // 添加主机
		hostRoute.GET("", h.listHosts)       // 获取主机列表
		hostRoute.GET("/:id", h.getHostById) // 根据id获取主机
		hostRoute.PUT("/:id", h.updateHost)  // 根据id修改主机
	}
}
