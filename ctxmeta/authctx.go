package ctxmeta

import (
	"encoding/json"
	"fmt"
	"strings"
)

type AuthType uint8 // 身份
const (
	AuthTypeUnknown  AuthType = 0 // 零值 未设置
	AuthTypeEmployee AuthType = 1 // 员工
	AuthTypeCustomer AuthType = 2 // 客户
	AuthTypeAgency   AuthType = 3 // 代理商
	// AuthTypeOpen     AuthType = 3 // 开放接口 如果我来设计开放接口也会是一个员工标识来着
)

func (a AuthType) String() string {
	return fmt.Sprintf("%d", int8(a))
}

// Employee 员工信息
type Employee struct {
	RoleIds         []int64 `json:"roleIds"`
	Username        string  `json:"username"`
	UserId          int64   `json:"userId"`
	PasswordVersion int64   `json:"passwordVersion"`
}

// Customer 客户信息
type Customer struct {
}

// Agency 代理商信息
type Agency struct {
}

// AuthCtx 授权信息 注意：设计上使用值语义
type AuthCtx struct {
	Type     AuthType `json:"type,omitempty"`     // admin_api 这里不大可能有其他身份，但是将来会有其他api 这里举个全例子
	Id       int64    `json:"id,omitempty"`       // 根据类型对应标识 员工标识 客户标识 开放平台标识
	Ip       string   `json:"ip,omitempty"`       // 来源IP地址
	ClientId string   `json:"cid,omitempty"`      // 客户端标识
	Employee Employee `json:"employee,omitempty"` // 员工信息
	Customer Customer `json:"customer,omitempty"` // 客户信息
	Agency   Agency   `json:"agency,omitempty"`   // 代理商信息
	ACL      ACL      `json:"acl,omitempty"`      // 可访问授权信息 数据权限 角色权限等
}

// IsEmployee 是否是员工
func (a AuthCtx) IsEmployee() bool {
	return a.Type == AuthTypeEmployee
}

// EmployeeId 员工标识，如果是员工就返回员工标识，不是员工返回0
func (a AuthCtx) EmployeeId() int64 {
	if a.IsEmployee() {
		return a.Id
	}
	return 0
}

// IsCustomer 是否是客户
func (a AuthCtx) IsCustomer() bool {
	return a.Type == AuthTypeCustomer
}

// CustomerId 客户标识，如果是客户就返回客户标识，不是返回0
func (a AuthCtx) CustomerId() int64 {
	if a.IsCustomer() {
		return a.Id
	}
	return 0
}

// IsAgency 是否是代理商
func (a AuthCtx) IsAgency() bool {
	return a.Type == AuthTypeAgency
}

// AgencyId 代理商标识
func (a AuthCtx) AgencyId() int64 {
	if a.IsAgency() {
		return a.Id
	}
	return 0
}

// Valid 授权信息是否正常
func (a AuthCtx) Valid() bool {
	return a.Type != AuthTypeUnknown
}

func (a AuthCtx) JSON() string {
	sb := &strings.Builder{}
	json.NewEncoder(sb).Encode(a)
	return sb.String()
}

func parseAuthCtx(s string) AuthCtx {
	var a AuthCtx
	json.NewDecoder(strings.NewReader(s)).Decode(&a)
	return a
}

// ACL 权限设计 注意：设计上使用值语义
type ACL struct {
	IsManager bool
}
