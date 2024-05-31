package layouts

type AppLayoutData struct {
	title string
	isDev bool
}

func (a *AppLayoutData) SetTitle(t string) {
	a.title = t
}

func (a AppLayoutData) GetTitle() string {
	return a.title
}

func (a *AppLayoutData) SetIsDev(b bool) {
	a.isDev = b
}

func (a AppLayoutData) GetIsDev() bool {
	return a.isDev
}
