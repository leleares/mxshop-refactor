package data

type DataFactory interface {
	Goods() GoodsData
	Users() UserData
}
