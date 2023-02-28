package web

//face info
type Web struct {
	app *App
	cookie *Cookie
	Base
}

//construct
func NewWeb() *Web {
	this := &Web{
		app: NewApp(),
		cookie: NewCookie(),
	}
	return this
}

func (f *Web) GetApp() *App {
	return f.app
}

func (f *Web) GetCookie() *Cookie {
	return f.cookie
}