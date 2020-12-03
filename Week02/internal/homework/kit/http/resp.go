package http

type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Resp struct {
	Status
	Data interface{} `json:"data"`
}
