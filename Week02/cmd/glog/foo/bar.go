package foo

import "github.com/golang/glog"

func FuncInBar() {
	glog.V(8).Info("LEVEL 8: level 8 message in funcInBar")
}
