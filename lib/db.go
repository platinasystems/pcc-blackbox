package pcc

type DBHandler struct {
}

type DBConfiguration struct {
	Address string `"json:address"`
	Port    int    `"json:port"`
	Name    string `"json:name"`
	User    string `"json:user"`
	Pwd     string `"json:pwd"`
}
