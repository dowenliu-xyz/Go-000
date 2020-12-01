package endpoint

import (
	"encoding/json"
	"fmt"
	"github.com/dowenliu-xyz/Go-000/Week02/homework/kit/errors"
	kithttp "github.com/dowenliu-xyz/Go-000/Week02/homework/kit/http"
	"github.com/dowenliu-xyz/Go-000/Week02/homework/staff/biz"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
	"strconv"
)

func writeResp(w io.Writer, resp interface{}) {
	bytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Json序列化失败, %v", err)
		return
	}
	_, _ = w.Write(bytes)
}

func GetStaff(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idStr := ps.ByName("id")
	if idStr == "" {
		writeResp(w, kithttp.Status{Code: 400, Message: "参数错误：缺少id"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeResp(w, kithttp.Status{Code: 400, Message: "参数错误：id必须是数字"})
		return
	}
	staff, err := biz.GetStaff(int64(id))
	if err != nil {
		if errors.IsErrResourceNotFound(err) {
			// 不打印日志，通常id没有找到，给客户端报找不到就可以了，打日志没有必要
			writeResp(w, kithttp.Status{Code: 404, Message: fmt.Sprintf("未找到id为 %d 的员工记录", id)})
			return
		}
		log.Printf("%+v", err)
		writeResp(w, kithttp.Status{Code: 500, Message: "服务器内部错误"})
		return
	}
	writeResp(w, kithttp.Resp{
		Status: kithttp.Status{Code: 200, Message: "ok"},
		Data:   staff,
	})
}
