package ui

var refreshRegistry = make(map[string]func())

func RegisterRefresh(tabName string, fn func()) {
	refreshRegistry[tabName] = fn
}

func GetRefresh(tabName string) func() {
	return refreshRegistry[tabName]
}
