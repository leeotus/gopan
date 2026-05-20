// package model 定义数据库表结构对应的 Go 结构体。
// 字段上的 db tag 用于 go-zero sqlx 的 ORM 映射。
package model

import (
	"database/sql"
	"time"
)

// User 映射 users 表，存储用户注册、登录及个人信息。
type User struct {
	Id        int64        `db:"id"`        // 主键，自增
	Username  string       `db:"username"`  // 用户名，唯一索引
	Password  string       `db:"password"`  // bcrypt 哈希后的密码
	Email     string       `db:"email"`     // 邮箱，唯一索引
	Avatar    string       `db:"avatar"`    // 头像 URL
	Signature string       `db:"signature"` // 个性签名
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"` // 软删除标记，NULL 表示未被删除
}
