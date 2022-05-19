package main

import "save-your-neck/service"

func main() {
	mw := service.NewWindow("save your neck")
	n := service.NewNoifier()
	n.MyWindow = mw
	go n.Run()
	mw.Show()
	mw.Run()
}
