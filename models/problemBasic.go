package models

import "gorm.io/gorm"

type ProblemBasic struct {
	gorm.Model
	Identity          string             `gorm:"column:identity;type:varchar(36);" json:"identity"`                 // 问题表的唯一标识
	Title             string             `gorm:"column:title;type:varchar(255);" json:"title"`                      // 文章标题
	Content           string             `gorm:"column:content;type:text;" json:"content"`                          // 文章正文
	MaxRuntime        int                `gorm:"column:max_runtime;type:int;" json:"max_runtime"`                   // 最大运行时间
	MaxMem            int                `gorm:"column:max_mem;type:int;" json:"max_mem"`                           // 最大运行内存
	PassNum           int64              `gorm:"column:pass_num; type:int;" json:"pass_num"`                        // 完成问题的个数
	SubmitNum         int64              `gorm:"column:submit_num; type:int;" json:"submit_num"`                    // 问题的提交次数
	ProblemCategories []*ProblemCategory `gorm:"foreignKey:problem_id;references:id" json:"problem_categories"`     // 关联问题分类表
	TestCases         []*TestCase        `gorm:"foreignKey:problem_identity;references:identity" json:"test_cases"` // 关联测试用例表
}

func (table *ProblemBasic) TableName() string {
	return "problem_basic"
}
