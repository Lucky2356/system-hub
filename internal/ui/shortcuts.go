package ui

var refreshRegistry = make(map[string]func())

func RegisterRefresh(tabName string, fn func()) {
	refreshRegistry[tabName] = fn
}

func GetRefresh(tabName string) func() {
	return refreshRegistry[tabName]
}

// closers holds cleanup callbacks (e.g. stopping background pollers) that run
// when the main window is closed.
var closers []func()

func RegisterCloser(fn func()) {
	closers = append(closers, fn)
}

func RunClosers() {
	for _, fn := range closers {
		if fn != nil {
			fn()
		}
	}
	closers = nil
}
