package apollo

import (
	"fmt"
	"time"
)

type App struct {
	Client *Client
	stop   chan bool
	change chan string
	err    chan error
}

func New(opts ...ConfigOptions) *App {
	client := NewClient(opts...)
	app := &App{
		Client: client,
		stop:   make(chan bool, 1),
		change: make(chan string),
		err:    make(chan error),
	}
	return app
}

func (app *App) Start() {
	err := app.Client.getConfig()
	if err != nil {
		fmt.Printf("获取apollo默认配置发生错误.[%v]\n", err)
	}
}

func (app *App) Stop() {
	app.stop <- true
}

func (app *App) Listener(namespaces ...string) {
	for _, namespace := range namespaces {
		_, _ = app.Client.GetNamespace(namespace, defaultNotificationID)
	}
	go app.startLongPoll()

	for {
		select {
		case n := <-app.change:
			fmt.Printf("namespace [%s] had be changed.\n", n)
		case <-app.err:
			app.Stop()
		}
	}
}

func (app *App) startLongPoll() {
	timer := time.NewTimer(app.Client.interval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			err := app.longPoll()
			if err != nil {
				app.err <- err
			} else {
				timer.Reset(app.Client.interval)
			}
			fmt.Printf("long polling execute.\n")
		case <-app.stop:
			fmt.Printf("[%s] stop listener", app.Client.appId)
			return
		}
	}
}

func (app *App) longPoll() error {
	ns := make([]Notification, 0)
	app.Client.data.Range(func(key, value interface{}) bool {
		if data, ok := value.(*NamespaceData); ok {
			ns = append(ns, Notification{
				Namespace:      key.(string),
				NotificationID: data.NotificationId,
			})
		}
		return true
	})
	if len(ns) == 0 {
		return nil
	}
	if notifications, err := app.Client.GetNotifications(ns); err == nil {
		for _, notification := range notifications {
			v, _ := app.Client.GetNamespace(notification.Namespace, notification.NotificationID)
			if v {
				app.change <- notification.Namespace
			}
		}
	}
	return nil
}
