package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/clientapp/models"
	"gin-fast/plugins/clientapp/service"

	"github.com/gin-gonic/gin"
)

// UserController 客户端用户控制器
type UserController struct {
	controllers.Common
	userService *service.UserService
}

func NewUserController() *UserController {
	return &UserController{
		userService: service.NewUserService(),
	}
}

func (uc *UserController) List(c *gin.Context) {
	var req models.UserListRequest
	if err := req.Validate(c); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	list, total, err := uc.userService.List(c, req)
	if err != nil {
		uc.FailAndAbort(c, "获取用户列表失败", err)
	}
	uc.Success(c, gin.H{"list": list, "total": total})
}

func (uc *UserController) GetByID(c *gin.Context) {
	var req models.UserGetByIDRequest
	if err := req.Validate(c); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	user, err := uc.userService.GetByID(c, req.ID)
	if err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	uc.Success(c, user)
}

func (uc *UserController) Create(c *gin.Context) {
	var req models.UserCreateRequest
	if err := req.Validate(c); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	user, err := uc.userService.Create(c, req)
	if err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	uc.Success(c, gin.H{"id": user.ID})
}

func (uc *UserController) Update(c *gin.Context) {
	var req models.UserUpdateRequest
	if err := req.Validate(c); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	if err := uc.userService.Update(c, req); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	uc.SuccessWithMessage(c, "更新成功")
}

func (uc *UserController) UpdateStatus(c *gin.Context) {
	var req models.UserStatusRequest
	if err := req.Validate(c); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	if err := uc.userService.UpdateStatus(c, req.ID, req.Status); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	uc.SuccessWithMessage(c, "状态更新成功")
}

func (uc *UserController) Delete(c *gin.Context) {
	var req models.UserDeleteRequest
	if err := req.Validate(c); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	if err := uc.userService.Delete(c, req.ID); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	uc.SuccessWithMessage(c, "删除成功")
}

func (uc *UserController) BindIdentity(c *gin.Context) {
	var req models.UserIdentityBindRequest
	if err := req.Validate(c); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	identity, err := uc.userService.BindIdentity(c, req)
	if err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	uc.Success(c, gin.H{"id": identity.ID})
}

func (uc *UserController) UnbindIdentity(c *gin.Context) {
	var req models.UserIdentityUnbindRequest
	if err := req.Validate(c); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	if err := uc.userService.UnbindIdentity(c, req.ID); err != nil {
		uc.FailAndAbort(c, err.Error(), err)
	}
	uc.SuccessWithMessage(c, "解绑成功")
}
