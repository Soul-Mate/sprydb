package mapper

type Custom interface {
	ReadFromDB([]byte) // 从数据库中读出的数据
	WriteToDB() []byte // 往数据库写入数据
}