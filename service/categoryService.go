package service

import (
	"log"
	"net/http"
	"oj/define"
	"oj/helper"
	"oj/models"
	"oj/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetGategoryList
// @Tags 管理员私有方法
// @Summary 问题列表
// @param Authorization header string true "Authorization"
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /admin/category-list [get]
func GetGategoryList(c *gin.Context) {
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("GetProblemList Page strconv Error:", err)
		return
	}
	page = (page - 1) * size

	var count int64
	keyword := c.Query("keyword")

	categoryList := make([]*models.CategoryBasic, 0)
	err = utils.DB.Model(new(models.CategoryBasic)).Where("name like ?", "%"+keyword+"%").
		Count(&count).Limit(size).Offset(page).Find(&categoryList).Error
	if err != nil {
		log.Println("GetGategoryList Error:" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取分类列表失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  categoryList,
			"count": count,
		},
	})
}

// CategoryCreate
// @Tags 管理员私有方法
// @Summary 分类创建
// @param Authorization header string true "Authorization"
// @param name formData string true "name"
// @param parentId formData string true "parentId"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /admin/category-create [post]
func CategoryCreate(c *gin.Context) {
	name := c.PostForm("name")
	parentId, _ := strconv.Atoi(c.PostForm("parentId"))

	category := &models.CategoryBasic{
		Identity: helper.GetUUID(),
		Name:     name,
		ParentId: parentId,
	}

	err := utils.DB.Create(category).Error
	if err != nil {
		log.Println("GategoryCreate Error:" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "创建分类失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "创建分类成功",
	})
}

// CategoryModify
// @Tags 管理员私有方法
// @Summary 分类修改
// @param Authorization header string true "Authorization"
// @param identity formData string true "identity"
// @param name formData string true "name"
// @param parentId formData string true "parentId"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /admin/category-modify [put]
func CategoryModify(c *gin.Context) {
	name := c.PostForm("name")
	parentId, _ := strconv.Atoi(c.PostForm("parentId"))
	identity := c.PostForm("identity")
	if name == "" || identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	category := &models.CategoryBasic{
		Identity: identity,
		Name:     name,
		ParentId: parentId,
	}

	err := utils.DB.Model(new(models.CategoryBasic)).Where("identity = ?", identity).Updates(category).Error
	if err != nil {
		log.Println("GategoryModify Error:" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "修改分类失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "修改分类成功",
	})
}

// CategoryDelete
// @Tags 管理员私有方法
// @Summary 分类删除
// @param Authorization header string true "Authorization"
// @param identity query string true "identity"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /admin/category-delete [delete]
func CategoryDelete(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	var cnt int64
	err := utils.DB.Model(new(models.ProblemCategory)).Where("category_id = (select id from category_basic where identity = ? limit 1)", identity).Error
	if err != nil {
		log.Println("GategoryDelete Error:" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取分类关联的问题失败",
		})
		return
	}
	if cnt > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取分类下面已存在问题，不可删除",
		})
		return
	}

	err = utils.DB.Where("identity = ?", identity).Delete(new(models.CategoryBasic)).Error
	if err != nil {
		log.Println("Delete GategoryBasic Error:" + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
