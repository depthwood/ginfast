package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/service"

	"github.com/gin-gonic/gin"
)

// ClientController 客户端管理控制器
type ClientController struct {
	controllers.Common
	clientService *service.ClientService
}

func NewClientController() *ClientController {
	return &ClientController{
		clientService: service.NewClientService(),
	}
}

func (cc *ClientController) List(c *gin.Context) {
	var req models.ClientListRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	list, total, err := cc.clientService.List(c, req)
	if err != nil {
		cc.FailAndAbort(c, "获取客户端列表失败", err)
	}
	cc.Success(c, gin.H{"list": list, "total": total})
}

func (cc *ClientController) GetByID(c *gin.Context) {
	var req models.ClientGetByIDRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	client, err := cc.clientService.GetByID(c, req.ID)
	if err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.Success(c, client)
}

func (cc *ClientController) Create(c *gin.Context) {
	var req models.ClientCreateRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	client, err := cc.clientService.Create(c, req)
	if err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.Success(c, gin.H{"id": client.ID})
}

func (cc *ClientController) Update(c *gin.Context) {
	var req models.ClientUpdateRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	if err := cc.clientService.Update(c, req); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.SuccessWithMessage(c, "更新成功")
}

func (cc *ClientController) UpdateStatus(c *gin.Context) {
	var req models.ClientStatusRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	if err := cc.clientService.UpdateStatus(c, req.ID, req.Status); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.SuccessWithMessage(c, "状态更新成功")
}

func (cc *ClientController) Delete(c *gin.Context) {
	var req models.ClientDeleteRequest
	if err := req.Validate(c); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	if err := cc.clientService.Delete(c, req.ID); err != nil {
		cc.FailAndAbort(c, err.Error(), err)
	}
	cc.SuccessWithMessage(c, "删除成功")
}

func (cc *ClientController) Options(c *gin.Context) {
	list, err := cc.clientService.ListAll(c)
	if err != nil {
		cc.FailAndAbort(c, "获取客户端选项失败", err)
	}
	cc.Success(c, gin.H{"list": list})
}
