##### 简介
一个定时提醒你起来溜达一下的应用，防止你的脖子挂掉，目前只能用在windows平台。

##### Build app

In the directory containing `main.go` run

	go build
	
To get rid of the cmd window, instead run

	go build -ldflags="-H windowsgui"

##### Run app
	
	save-your-neck.exe