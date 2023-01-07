package service

import (
	"encoding/json"
	"errors"
	"log"
	"oj/dao"
	"oj/define"
	"oj/helper"
	"oj/models"
	"oj/utils"
	"strconv"

	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetProblemList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Param category_identity query string false "category_identity"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /problem-list [get]
func GetProblemList(c *gin.Context) {
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
	keyword := c.Query("keyword")
	categoryIdentity := c.Query("category_identity")

	list := make([]models.ProblemBasic, 0)
	page = (page - 1) * size
	var count int64

	tx := dao.GetProblemList(keyword, categoryIdentity)
	err = tx.Count(&count).Omit("content").Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		log.Println("Get Problem List Error:", err)
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

// GetProblemDetail
// @Tags 公共方法
// @Summary 问题详情
// @Param identity query string false "identity"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /problem-detail [get]
func GetProblemDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "问题唯一标识不能为空",
		})
		return
	}

	data := new(models.ProblemBasic)
	err := utils.DB.Where("identity = ?", identity).
		Preload("ProblemCategories").Preload("ProblemCategories.CategoryBasic").
		First(&data).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "问题不存在",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get ProblemDetail Error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// ProblemsCreate
// @Tags 管理员私有方法
// @Summary 问题创建
// @param Authorization header string true "Authorization"
// @param title formData string true "title"
// @param content formData string true "content"
// @param max_runtime formData string false "max_runtime"
// @param max_mem formData string false "max_mem"
// @param category_ids formData []string true "category_ids" collectionFormat(multi)
// @param test_cases formData []string true "test_cases" collectionFormat(multi)
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /admin/problem-create [post]
func ProblemCreate(c *gin.Context) {
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMemory, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")
	if title == "" || content == "" || len(categoryIds) == 0 || len(testCases) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}

	identity := helper.GetUUID()
	data := &models.ProblemBasic{
		Identity:   identity,
		Title:      title,
		Content:    content,
		MaxRuntime: maxRuntime,
		MaxMem:     maxMemory,
	}

	// 处理分类
	categoryBasics := make([]*models.ProblemCategory, 0)
	for _, id := range categoryIds {
		categoryId, _ := strconv.Atoi(id)
		categoryBasics = append(categoryBasics, &models.ProblemCategory{
			ProblemId:  data.ID,
			CategoryId: uint(categoryId),
		})
	}
	data.ProblemCategories = categoryBasics

	// 处理测试案例
	testCaseBasics := make([]*models.TestCase, 0)
	for _, testCase := range testCases {
		caseMap := make(map[string]string)
		err := json.Unmarshal([]byte(testCase), &caseMap)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}
		if _, ok := caseMap["input"]; !ok {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}
		if _, ok := caseMap["output"]; !ok {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}

		testCaseBasic := &models.TestCase{
			Identity:        helper.GetUUID(),
			ProblemIdentity: identity,
			Input:           caseMap["input"],
			Output:          caseMap["output"],
		}
		testCaseBasics = append(testCaseBasics, testCaseBasic)
	}
	data.TestCases = testCaseBasics

	// 创建问题
	err := utils.DB.Create(data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Problem Create Error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"identity": data.Identity,
		},
	})
}

// ProblemModify
// @Tags 管理员私有方法
// @Summary 问题修改
// @param Authorization header string true "Authorization"
// @param identity formData string true "identity"
// @param title formData string true "title"
// @param content formData string true "content"
// @param max_runtime formData string false "max_runtime"
// @param max_mem formData string false "max_mem"
// @param category_ids formData []string true "category_ids" collectionFormat(multi)
// @param test_cases formData []string true "test_cases" collectionFormat(multi)
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /admin/problem-modify [put]
func ProblemModify(c *gin.Context) {
	identity := c.PostForm("identity")
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMemory, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")
	if identity == "" || title == "" || content == "" || len(categoryIds) == 0 || len(testCases) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}

	if err := utils.DB.Transaction(func(tx *gorm.DB) error {
		// 问题基础信息保存 problem_basic
		problemBasic := &models.ProblemBasic{
			Identity:   identity,
			Title:      title,
			Content:    content,
			MaxRuntime: maxRuntime,
			MaxMem:     maxMemory,
		}
		err := tx.Where("identity = ?", identity).Updates(problemBasic).Error
		if err != nil {
			return err
		}

		// 查询问题详情
		err = tx.Where("identity = ?", identity).Find(&problemBasic).Error
		if err != nil {
			return err
		}

		// 关联问题分类的更新
		// 1、删除已存在的关联关系
		err = tx.Where("problem_id = ?", problemBasic.ID).Delete(new(models.ProblemCategory)).Error
		if err != nil {
			return err
		}
		// 2、新增新的关联关系
		pcs := make([]*models.ProblemCategory, 0)
		for _, id := range categoryIds {
			intId, _ := strconv.Atoi(id)
			pcs = append(pcs, &models.ProblemCategory{
				ProblemId:  problemBasic.ID,
				CategoryId: uint(intId),
			})
		}
		err = tx.Create(&pcs).Error
		if err != nil {
			return err
		}

		// 关联调试案例的更新
		// 1、删除已存在的关联关系
		err = tx.Where("problem_identity = ?", identity).Delete(new(models.TestCase)).Error
		if err != nil {
			return err
		}
		// 2、增加新的关联关系
		tcs := make([]*models.TestCase, 0)
		for _, testCase := range testCases {
			caseMap := make(map[string]string)
			err := json.Unmarshal([]byte(testCase), &caseMap)
			if err != nil {
				return err
			}

			if _, ok := caseMap["input"]; !ok {
				return errors.New("测试案例[input]格式错误")
			}
			if _, ok := caseMap["output"]; !ok {
				return errors.New("测试案例[output]格式错误")
			}
			tcs = append(tcs, &models.TestCase{
				Identity:        helper.GetUUID(),
				ProblemIdentity: identity,
				Input:           caseMap["input"],
				Output:          caseMap["output"],
			})
		}

		err = tx.Create(tcs).Error
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Problem Modify Error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "问题修改成功",
	})
}
