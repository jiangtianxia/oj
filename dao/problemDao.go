package dao

import (
	"oj/models"
	"oj/utils"

	"gorm.io/gorm"
)

func GetProblemList(keyword, categoryIdentity string) *gorm.DB {

	tx := utils.DB.Model(new(models.ProblemBasic)).Preload("ProblemCategories").Preload("ProblemCategories.CategoryBasic").
		Where("title like ? or content like ?", "%"+keyword+"%", "%"+keyword+"%")

	if categoryIdentity != "" {
		tx.Joins("RIGHT JOIN problem_category pc on pc.problem_id = problem_basic.id").
			Where("pc.category_id = (SELECT cb.id FROM category_basic cb WHERE cb.identity = ?)", categoryIdentity)
	}

	return tx
}
