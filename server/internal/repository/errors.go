// Package repository 是数据访问层：在 gorm-gen 生成的类型安全查询之上，
// 封装事务、乐观锁、软删除/回收站语义，并向上层暴露语义化错误。本层不含业务规则。
package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// 语义化仓储错误：上层据此翻译为领域错误，不感知 gorm/驱动细节。
var (
	ErrNotFound  = errors.New("repository: record not found")
	ErrConflict  = errors.New("repository: optimistic lock conflict")
	ErrDuplicate = errors.New("repository: duplicate key")
)

// mapError 将 gorm / 驱动错误翻译为语义化仓储错误。
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicate
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == 1062 { // ER_DUP_ENTRY
		return ErrDuplicate
	}
	return err
}
