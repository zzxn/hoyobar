package handler

type UserHelloReq struct {
    Msg string `json:"msg"`
}

type UserRegisterReq struct {
    Username string
    Password string
    Vcode string
}
