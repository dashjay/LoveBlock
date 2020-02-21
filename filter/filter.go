package filter

import "github.com/importcjj/sensitive"

var filter *sensitive.Filter

func init() {
	//filter = sensitive.New()
	//err := filter.LoadWordDict("./dic.txt")
	//if err != nil {
	//	panic(err)
	//}
}

func GetFilter() *sensitive.Filter {
	return filter
}
