package service

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-toast/toast"
)

type Notifier struct {
	CycleMinutes int //cycle miniutes
	DelayMinutes int
	Cur          int
	Goal         int
	IsRunning    bool
	GinPort      string
	OKURL        string
	DelayURL     string
	MyWindow     *MyWindow
}

func NewNoifier() (n *Notifier) {
	n = &Notifier{
		CycleMinutes: 45,
		DelayMinutes: 15,
		GinPort:      ":8866",
	}
	n.OKURL = "http://localhost" + n.GinPort + "/notify?reply=ok"
	n.DelayURL = "http://localhost" + n.GinPort + "/notify?reply=delay"
	return n
}

func (n *Notifier) Run() {
	n.IsRunning = true
	n.Goal = n.CycleMinutes
	go func() {
		for n.IsRunning {
			n.Cur++
			time.Sleep(time.Minute)
			if n.Cur > n.Goal {
				n.Goal = n.CycleMinutes
				n.Cur = 0
				n.Push()
			}
		}
	}()
	n.Serve()
}
func (n *Notifier) Stop() {
	n.IsRunning = false
}

func (n *Notifier) DealDelay() {
	n.Cur = 0
	n.Goal = n.DelayMinutes
}

func (n *Notifier) DealOK() {
	n.Cur = 0
	n.Goal = n.DelayMinutes + n.CycleMinutes
}

func (n *Notifier) Push() {
	notification := toast.Notification{
		AppID:   "SaveYourNeck",
		Title:   "It is time to save your neck!!",
		Message: "Ok or die!!!",
		Actions: []toast.Action{
			{"protocol", "ok", n.OKURL},
			{"protocol", "delay", n.DelayURL},
		},
	}
	n.MyWindow.AddLog("push")
	notification.Push()
}

func (n *Notifier) Serve() {
	// 1.创建路由
	r := gin.Default()
	// 2.绑定路由规则，执行的函数
	// gin.Context，封装了request和response
	r.GET("/notify", func(c *gin.Context) {
		r := c.Query("reply")
		if r == "ok" {
			n.MyWindow.AddLog("reply ok")
			n.DealOK()
		}
		if r == "delay" {
			n.MyWindow.AddLog("reply delay")
			n.DealDelay()
		}
		c.JSON(200, gin.H{
			"message": r,
		})
	})
	// 3.监听端口，默认在8080
	// Run("里面不指定端口号默认为8080")
	r.Run(n.GinPort)
}
