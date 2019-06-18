package gui

type GUI interface {
	Start(windowTitle string) error
	Terminate() error
	SendTextMessage(text string) error
	SendErrorMessage(err error) error
	SendCloseMessage() error
	SetTitle(title string) error
	SetProgressMax(val int)
	ProgressStep()
	Wait()
	Closed() bool
}

//go:generate go-bindata-assetfs -pkg gui assets/...
