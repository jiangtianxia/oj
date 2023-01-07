package middlewares

import (
	"net/http"
	"oj/helper"

	"github.com/gin-gonic/gin"
)

// 验证是否为用户
func AuthUserCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		userClaim, err := helper.AnalyseToken(auth)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Unauthorized Authorization",
			})
			return
		}

		if userClaim == nil {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Unauthorized Admin",
			})
			return
		}
		c.Set("user", userClaim)
		c.Next()
	}
}
