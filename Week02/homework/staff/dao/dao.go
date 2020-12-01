package dao

import (
	"database/sql"
	kiterr "github.com/dowenliu-xyz/Go-000/Week02/homework/kit/errors"
	"github.com/dowenliu-xyz/Go-000/Week02/homework/staff"
	"github.com/pkg/errors"
)

var mockStorage map[int64]*staff.Staff

func init() {
	mockStorage = map[int64]*staff.Staff{
		1: {1, "小A"},
		2: {2, "小B"},
	}
}

// StaffById 模拟从数据源中取出员工的操作，如果找到 id 对应的员工，返回员工信息和一个nil error。
// 如果未打到，返回 nil 和一个包装过的 ErrResourceNotFound 类型
func StaffById(id int64) (*staff.Staff, error) {
	// 如果 id 小于，制造一个 panic。注意，此panic 仅为示例验证panic兜底使用，工作代码绝对不能用 panic 来抛错
	if id < 0 {
		panic(errors.New("Oh, 出错了，与数据源失去连接"))
	}
	// 如果 id 为 0 ，正常抛出一个连接已关闭错误
	if id == 0 {
		return nil, errors.New("Connection lost.")
	}
	if aStaff, ok := mockStorage[id]; ok {
		return aStaff, nil
	}
	err := kiterr.NewErrResourceNotFound(sql.ErrNoRows)
	return nil, errors.Wrapf(err, "数据源中未找到id为 %d 的员工", id)
}
