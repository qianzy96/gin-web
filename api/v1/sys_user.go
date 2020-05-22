package v1

import (
	"gin-web/models"
	"gin-web/pkg/cache_service"
	"gin-web/pkg/global"
	"gin-web/pkg/request"
	"gin-web/pkg/response"
	"gin-web/pkg/service"
	"gin-web/pkg/utils"
	"github.com/gin-gonic/gin"
)

// 获取当前用户信息
func GetUserInfo(c *gin.Context) {
	user := GetCurrentUser(c)
	// 转为UserInfoResponseStruct, 隐藏部分字段
	var resp response.UserInfoResponseStruct
	utils.Struct2StructByJson(user, &resp)
	resp.Roles = []string{
		"admin",
	}
	resp.Permissions = []string{
		"***",
	}
	response.SuccessWithData(resp)
}

// 获取用户列表
func GetUsers(c *gin.Context) {
	// 绑定参数
	var req request.UserListRequestStruct
	_ = c.Bind(&req)
	// 创建服务
	s := cache_service.New(c)
	users, err := s.GetUsers(&req)
	if err != nil {
		response.FailWithMsg(err.Error())
		return
	}
	// 转为ResponseStruct, 隐藏部分字段
	var respStruct []response.UserListResponseStruct
	utils.Struct2StructByJson(users, &respStruct)
	// 返回分页数据
	var resp response.PageData
	// 设置分页参数
	resp.PageInfo = req.PageInfo
	// 设置数据列表
	resp.List = respStruct
	response.SuccessWithData(resp)
}

// 修改密码
func ChangePwd(c *gin.Context) {
	var msg string
	// 请求json绑定
	var req request.ChangePwdRequestStruct
	_ = c.ShouldBindJSON(&req)
	// 获取当前用户
	user := GetCurrentUser(c)
	query := global.Mysql.Where("username = ?", user.Username).First(&user)
	// 查询用户
	err := query.Error
	if err != nil {
		msg = err.Error()
	} else {
		// 校验密码
		if ok := utils.ComparePwd(req.OldPassword, user.Password); !ok {
			msg = "原密码错误"
		} else {
			// 更新密码
			err = query.Update("password", utils.GenPwd(req.NewPassword)).Error
			if err != nil {
				msg = err.Error()
			}
		}
	}
	if msg != "" {
		response.FailWithMsg(msg)
		return
	}
	response.Success()
}

// 获取当前请求用户信息
func GetCurrentUser(c *gin.Context) models.SysUser {
	user, _ := c.Get("user")
	u, _ := user.(models.SysUser)
	// 创建服务
	s := cache_service.New(c)
	newUser, _ := s.GetUserById(u.Id)
	return newUser
}

// 创建用户
func CreateUser(c *gin.Context) {
	user := GetCurrentUser(c)
	// 绑定参数
	var req request.CreateUserRequestStruct
	_ = c.Bind(&req)
	// 参数校验
	err := global.NewValidatorError(global.Validate.Struct(req), req.FieldTrans())
	if err != nil {
		response.FailWithMsg(err.Error())
		return
	}
	// 记录当前创建人信息
	req.Creator = user.Nickname + user.Username
	// 创建服务
	s := service.New(c)
	err = s.CreateUser(&req)
	if err != nil {
		response.FailWithMsg(err.Error())
		return
	}
	response.Success()
}

// 更新用户
func UpdateUserById(c *gin.Context) {
	// 绑定参数
	var req gin.H
	var pwd request.ChangePwdRequestStruct
	_ = c.Bind(&req)
	// 将部分参数转为pwd, 如果值不为空, 可能会用到
	utils.Struct2StructByJson(req, &pwd)
	// 获取path中的userId
	userId := utils.Str2Uint(c.Param("userId"))
	if userId == 0 {
		response.FailWithMsg("用户编号不正确")
		return
	}
	// 创建服务
	s := service.New(c)
	// 更新数据
	err := s.UpdateUserById(userId, pwd.NewPassword, req)
	if err != nil {
		response.FailWithMsg(err.Error())
		return
	}
	response.Success()
}

// 批量删除用户
func BatchDeleteUserByIds(c *gin.Context) {
	var req request.Req
	_ = c.Bind(&req)
	// 创建服务
	s := service.New(c)
	// 删除数据
	err := s.DeleteUserByIds(req.GetUintIds())
	if err != nil {
		response.FailWithMsg(err.Error())
		return
	}
	response.Success()
}
