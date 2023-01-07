package service

import (
	"log"
	"net/http"
	"oj/define"
	"oj/helper"
	"oj/models"
	"oj/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserDetail
// @Tags 公共方法
// @Summary 用户详情
// @Param identity query string false "identity"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /user-detail [get]
func GetUserDetail(c *gin.Context) {
	identity := c.Query("identity")

	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "用户唯一标识不能为空",
		})
		return
	}

	data := new(models.UserBasic)
	err := utils.DB.Omit("password").Where("identity = ?", identity).Find(&data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get User Detail By Identity:" + identity + "Error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// Login
// @Tags 公共方法
// @Summary 用户登录
// @param username formData string false "username"
// @param password formData string false "password"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /login [post]
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "必填信息为空",
		})
		return
	}
	password = helper.GetMd5(password)

	data := new(models.UserBasic)
	err := utils.DB.Where("name = ? AND password = ?", username, password).First(&data).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "用户名或密码错误",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get UserBasic Error：" + err.Error(),
		})
		return
	}

	token, err := helper.GenerateToken(data.Identity, data.Name, data.IsAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get GenerateTokenc Error：" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// SendCode
// @Tags 公共方法
// @Summary 发送验证码
// @param email formData string true "email"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /send-code [post]
func SendCode(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	code := helper.GetRand()
	utils.RDB.Set(c, email, code, time.Second*300)
	err := helper.SendCode(email, code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Send Code Error：" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "验证码发送成功",
	})
}

// Register
// @Tags 公共方法
// @Summary 用户注册
// @param mail formData string true "mail"
// @param code formData string true "code"
// @param name formData string true "name"
// @param password formData string true "password"
// @param phone formData string false "phone"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /register [post]
func Register(c *gin.Context) {
	mail := c.PostForm("mail")
	userCode := c.PostForm("code")
	name := c.PostForm("name")
	password := c.PostForm("password")
	phone := c.PostForm("phone")
	if mail == "" || userCode == "" || name == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	// 验证验证码是否正确
	sysCode, err := utils.RDB.Get(c, mail).Result()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "验证码已过期，请重新获取验证码",
		})
		log.Println("Get Code Error：", err.Error())
		return
	}
	if sysCode != userCode {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "验证码不正确",
		})
		return
	}

	// 判断邮箱是否已存在
	var cnt int64
	err = utils.DB.Where("mail = ?", mail).Model(new(models.UserBasic)).Count(&cnt).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get User Error：" + err.Error(),
		})
		return
	}

	if cnt > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "该邮箱已被注册",
		})
		return
	}

	// 数据的插入
	userIdentity := helper.GetUUID()
	data := &models.UserBasic{
		Identity: userIdentity,
		Name:     name,
		Password: helper.GetMd5(password),
		Phone:    phone,
		Mail:     mail,
	}
	err = utils.DB.Create(&data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Creat User Error:" + err.Error(),
		})
		return
	}

	// 生成token
	token, err := helper.GenerateToken(userIdentity, name, data.IsAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Generate Token Error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "注册成功",
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// GetRankList
// @Tags 公共方法
// @Summary 用户排行榜
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /rank-list [get]
func GetRankList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("GetProblem Page strconv Error:", err)
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	if err != nil {
		log.Println("GetProblem Size strconv Error:", err)
		return
	}

	list := make([]models.UserBasic, 0)
	page = (page - 1) * size
	var count int64

	err = utils.DB.Model(new(models.UserBasic)).Count(&count).Order("finish_problem_num DESC, submit_num ASC").
		Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Gen Rank List Error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  list,
			"count": count,
		},
	})
}
