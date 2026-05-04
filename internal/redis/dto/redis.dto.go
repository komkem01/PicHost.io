package redisdto

type Option struct {
	Addr     string
	Db       int
	Username string
	Password string
	TimeZone string
}

func (rOpt *Option) DB() int {
	return rOpt.Db
}
