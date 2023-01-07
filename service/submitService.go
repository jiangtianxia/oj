package service

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"oj/define"
	"oj/helper"
	"oj/models"
	"oj/utils"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// GetSubmitList
// @Tags 公共方法
// @Summary 提交列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Param problem_identity query string false "problem_identity"
// @Param user_identity query string false "user_identity"
// @Param status query int false "status"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /submit-list [get]
func GetSubmitList(c *gin.Context) {
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("GetProblemList Page strconv Error:", err)
		return
	}
	page = (page - 1) * size

	var count int64
	list := make([]models.SubmitBasic, 0)

	problemIdentity := c.Query("problem_identity")
	userIdentity := c.Query("user_identity")
	status, _ := strconv.Atoi(c.Query("status"))
	tx := models.GetSubmitList(problemIdentity, userIdentity, status)
	err = tx.Count(&count).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		log.Println("Get Problem List Error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get Submit List Error:" + err.Error(),
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

// Submit
// @Tags 用户私有方法
// @Summary 代码提交
// @param Authorization header string true "Authorization"
// @Param problem_identity query string true "problem_identity"
// @Param code body string true "code"
// @Success 200 {string} json "{"code":"200", "data":""}"
// @Router /user/submit [post]
func Submit(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	code, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("Read Code Error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Read Code Error:" + err.Error(),
		})
		return
	}

	// 代码保存
	path, err := helper.CodeSave(code)
	if err != nil {
		log.Println("Code Save Error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Code Save Error:" + err.Error(),
		})
		return
	}

	// 提交
	u, _ := c.Get("user")
	userClaim := u.(*helper.UserClaims)
	sb := &models.SubmitBasic{
		Identity:        helper.GetUUID(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Path:            path,
	}

	// 代码判断
	pb := new(models.ProblemBasic)
	err = utils.DB.Where("identity = ?", problemIdentity).Preload("TestCases").First(pb).Error
	if err != nil {
		log.Println("Get Problem Error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get Problem Error:" + err.Error(),
		})
		return
	}

	// 提交判断
	// 答案错误的channel
	WA := make(chan int)
	// 超内存的channel
	OOM := make(chan int)
	// 编译错误的channer
	CE := make(chan int)
	// 通过的个数
	passCount := 0
	var lock sync.Mutex
	// 提示信息
	var msg string

	for _, testCase := range pb.TestCases {
		testCase := testCase
		go func() {
			// 执行测试
			wd, _ := os.Getwd()
			// 输出目录，看看路径对不对
			fmt.Println("工作目录: " + wd)
			fmt.Println("PATH:", path)
			fmt.Println(wd + "/" + path)
			cmd := exec.Command("go", "run", wd+"/"+path)
			var out, stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &out

			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				log.Fatalln(err)
			}
			io.WriteString(stdinPipe, testCase.Input)

			// 根据测试的输入案例进行运行，拿到输出结果和标准的输出结果进行比对
			var bm runtime.MemStats
			runtime.ReadMemStats(&bm)
			if err := cmd.Run(); err != nil {
				log.Println(err, stderr.String())
				if err.Error() == "exit status 2" {

					msg = stderr.String()
					CE <- 1
					return
				}
			}
			var em runtime.MemStats
			runtime.ReadMemStats(&em)
			fmt.Println(out.String())
			// 答案错误
			if testCase.Output != out.String() {
				msg = "答案错误"
				WA <- 1
				return
			}

			// 运行超内存
			if em.Alloc/1024-(bm.Alloc/1024) > uint64(pb.MaxMem) {
				msg = "运行超内存"
				OOM <- 1
				return
			}
			lock.Lock()
			passCount++
			lock.Unlock()
		}()
	}

	select {
	case <-WA:
		sb.Status = 2
	case <-OOM:
		sb.Status = 4
	case <-time.After(time.Microsecond * time.Duration(pb.MaxRuntime)):
		if passCount == len(pb.TestCases) {
			sb.Status = 1
		} else {
			sb.Status = 3
		}
	case <-CE:
		sb.Status = 5
	}

	err = utils.DB.Create(sb).Error
	if err != nil {
		log.Println("SubmitBasic Create Error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "SubmitBasic Create Error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"status": sb.Status,
			"msg":    msg,
		},
	})
}
