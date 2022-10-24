package host

import (
	"github.com/gin-gonic/gin"
	"strconv"

	"github.com/caoyingjunz/gopixiu/api/server/httpstatus"
	"github.com/caoyingjunz/gopixiu/api/server/httputils"
	"github.com/caoyingjunz/gopixiu/pkg/db/model"
	"github.com/caoyingjunz/gopixiu/pkg/pixiu"
)

// @Summary      Add host
// @Description  Add host
// @Tags         host
// @Accept       json
// @Produce      json
// @Param        data body model.MachineReq true "host info"
// @Success      200  {object}  httputils.HttpOK
// @Failure      400  {object}  httputils.HttpError
// @Router       /hosts [post]
func (h *hostRouter) addHost(c *gin.Context) {

	r := httputils.NewResponse()

	var machine model.MachineReq
	if err := c.ShouldBindJSON(&machine); err != nil {
		httputils.SetFailed(c, r, httpstatus.ParamsError)
		return
	}

	if err := pixiu.CoreV1.Machine().Create(c, &machine); err != nil {
		httputils.SetFailed(c, r, httpstatus.OperateFailed)
		return
	}

	httputils.SetSuccess(c, r)

}

// @Summary      List host
// @Description  List host
// @Tags         host
// @Accept       json
// @Produce      json
// @Success      200  {object}  httputils.Response{result=model.Machine}
// @Failure      400  {object}  httputils.HttpError
// @Router       /hosts [get]
func (h *hostRouter) listHosts(c *gin.Context) {

	r := httputils.NewResponse()

	res, err := pixiu.CoreV1.Machine().List(c)
	if err != nil {
		httputils.SetFailed(c, r, httpstatus.OperateFailed)
		return
	}

	r.Result = res

	httputils.SetSuccess(c, r)
}

// @Summary      Get host by id
// @Description  Get host by id
// @Tags         host
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "host ID"
// @Success      200  {object}  httputils.Response{result=model.Machine}
// @Failure      400  {object}  httputils.HttpError
// @Router       /hosts/:id [get]
func (h *hostRouter) getHostById(c *gin.Context) {

	r := httputils.NewResponse()

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		httputils.SetFailed(c, r, httpstatus.ParamsError)
		return
	}
	res, err := pixiu.CoreV1.Machine().GetByMachineId(c, id)
	if err != nil {
		httputils.SetFailed(c, r, httpstatus.OperateFailed)
		return
	}

	r.Result = res

	httputils.SetSuccess(c, r)
}

// @Summary      Get host by id
// @Description  Get host by id
// @Tags         host
// @Accept       json
// @Produce      json
// @Param        data body model.MachineReq true "host info"
// @Param        id   path      int  true  "host ID"
// @Success      200  {object}  httputils.Response{result=model.Machine}
// @Failure      400  {object}  httputils.HttpError
// @Router       /hosts/:id [put]
func (h *hostRouter) updateHost(c *gin.Context) {

	r := httputils.NewResponse()

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		httputils.SetFailed(c, r, httpstatus.ParamsError)
		return
	}

	var machine model.MachineReq
	if err := c.ShouldBindJSON(&machine); err != nil {
		httputils.SetFailed(c, r, httpstatus.ParamsError)
		return
	}

	if !pixiu.CoreV1.Machine().CheckMachineExist(c, id) {
		httputils.SetFailed(c, r, httpstatus.OperateFailed)
		return
	}

	if err := pixiu.CoreV1.Machine().Update(c, id, &machine); err != nil {
		httputils.SetFailed(c, r, httpstatus.OperateFailed)
		return
	}

	httputils.SetSuccess(c, r)
}
