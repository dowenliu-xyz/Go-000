package biz

import (
	"github.com/dowenliu-xyz/Go-000/Week02/homework/staff"
	"github.com/dowenliu-xyz/Go-000/Week02/homework/staff/dao"
	"github.com/pkg/errors"
)

func GetStaff(id int64) (*staff.Staff, error) {
	staffById, err := dao.StaffById(id)
	if err != nil {
		return nil, errors.WithMessage(err, "未取得指定员工数据")
	}
	return staffById, nil
}
