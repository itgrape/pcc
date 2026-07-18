// Command gen 使用 gorm-gen 从 model 结构体生成类型安全的查询代码（DAO/DAL）到 internal/query。
// 结构体优先：不连接数据库、不执行任何迁移；SQL 由 DBA 单独维护，二者手工对齐。
package main

import (
	"gorm.io/gen"

	"github.com/itgrape/pcc/server/internal/model"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:       "./internal/query",
		ModelPkgPath:  "./internal/model",
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable: true,
	})

	g.ApplyBasic(
		model.Namespace{},
		model.Config{},
		model.ConfigHistory{},
		model.AccessKey{},
		model.AccessKeyNamespace{},
	)

	g.Execute()
}
