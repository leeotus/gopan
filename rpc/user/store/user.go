// package store 封装 users 表的所有数据库 CRUD 操作。
// 使用 go-zero 自带的 sqlx 作为轻量 ORM，直接写 SQL。
// 所有方法接收 context.Context 以支持超时控制和链路追踪。
package store

import (
	"context"
	"database/sql"

	"gopan/rpc/user/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UserStore 封装 users 表的数据访问。
type UserStore struct {
	conn sqlx.SqlConn // go-zero sqlx 连接，支持 MySQL
}

// NewUserStore 创建 UserStore 实例。
func NewUserStore(conn sqlx.SqlConn) *UserStore {
	return &UserStore{conn: conn}
}

// Insert 插入一条用户记录。
// 返回的 sql.Result 可通过 LastInsertId() 获取自增 ID。
func (s *UserStore) Insert(ctx context.Context, u *model.User) (sql.Result, error) {
	return s.conn.ExecCtx(ctx, `
		INSERT INTO users (username, password, email, avatar, signature, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`, u.Username, u.Password, u.Email, u.Avatar, u.Signature)
}

// FindByUsername 根据用户名查找用户（用于登录校验和重名检测）。
// 返回 sql.ErrNoRows 表示用户不存在。
func (s *UserStore) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := s.conn.QueryRowCtx(ctx, &u, `
		SELECT id, username, password, email, avatar, signature, created_at, updated_at
		FROM users WHERE username = ? AND deleted_at IS NULL
	`, username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindById 根据主键 ID 查找用户。
func (s *UserStore) FindById(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := s.conn.QueryRowCtx(ctx, &u, `
		SELECT id, username, password, email, avatar, signature, created_at, updated_at
		FROM users WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Update 更新用户资料（邮箱、头像、签名），不更新用户名和密码。
// 仅当 deleted_at IS NULL 时执行。
func (s *UserStore) Update(ctx context.Context, u *model.User) error {
	_, err := s.conn.ExecCtx(ctx, `
		UPDATE users SET email = ?, avatar = ?, signature = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, u.Email, u.Avatar, u.Signature, u.Id)
	return err
}
