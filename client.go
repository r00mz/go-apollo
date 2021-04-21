package apollo

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
)

type NamespaceData struct {
	NotificationId int
	Configurations map[string]interface{}
	ReleaseKey     string
}

type Notification struct {
	Namespace      string `json:"namespaceName"`
	NotificationID int    `json:"notificationId"`
}

type Client struct {
	Config
	data   sync.Map
	isInit atomic.Value
}

func NewClient(opts ...ConfigOptions) *Client {
	c := &Client{
		Config: NewConfig(opts...),
	}
	c.isInit.Store(false)
	return c
}

func (c *Client) GetString(key string) (string, error) {
	if !c.isInit.Load().(bool) {
		err := c.getConfig()
		if err != nil {
			return "", err
		}
	}
	if value, ok := c.data.Load(c.namespaces); ok {
		if data, ok := value.(*NamespaceData); ok {
			return data.Configurations[key].(string), nil
		}
	}
	return "", errors.New("can not find the data[" + key + "]")
}

func (c *Client) getConfig() error {
	serverUrl := fmt.Sprintf("%s/configfiles/json/%s/%s/%s?ip=%s",
		c.server,
		url.QueryEscape(c.appId),
		url.QueryEscape(c.cluster),
		url.QueryEscape(c.namespaces),
		c.clientIp,
	)
	var res map[string]interface{}
	status, err := HttpGet(serverUrl, 3, &res)

	if err == nil && status == 200 {
		c.data.Store(c.namespaces, &NamespaceData{
			NotificationId: defaultNotificationID,
			ReleaseKey:     "",
			Configurations: res,
		})
		if !c.isInit.Load().(bool) {
			c.isInit.Store(true)
		}
		return err
	}
	return errors.New("load config fail")
}

func (c *Client) GetNamespace(namespace string, notificationId int) (bool, error) {
	if namespace == "" {
		namespace = defaultNamespaces
	}
	releaseKey := ""

	if value, ok := c.data.Load(namespace); ok {
		data := value.(*NamespaceData)
		releaseKey = data.ReleaseKey
		if data.NotificationId != notificationId {
			data.NotificationId = notificationId
			c.data.Store(namespace, data)
		}
	}

	if notificationId == 0 {
		notificationId = defaultNotificationID
	}

	serverUrl := fmt.Sprintf("%s/configs/%s/%s/%s?releaseKey=%s&ip=%s",
		c.server,
		url.QueryEscape(c.appId),
		url.QueryEscape(c.cluster),
		url.QueryEscape(namespace),
		url.QueryEscape(releaseKey),
		c.clientIp,
	)

	var res struct {
		AppID          string                 `json:"appId"`
		Cluster        string                 `json:"cluster"`
		Name           string                 `json:"namespaceName"`
		Configurations map[string]interface{} `json:"configurations"`
		ReleaseKey     string                 `json:"releaseKey"`
	}
	status, err := HttpGet(serverUrl, 3, &res)
	if err == nil && status == 200 {
		c.data.Store(namespace, &NamespaceData{
			NotificationId: notificationId,
			ReleaseKey:     res.ReleaseKey,
			Configurations: res.Configurations,
		})
		if !c.isInit.Load().(bool) {
			c.isInit.Store(true)
		}
		return true, nil
	}

	return false, err
}

func (c *Client) GetNotifications(ns []Notification) ([]Notification, error) {
	if ns == nil || len(ns) == 0 {
		ns = append(ns, Notification{Namespace: defaultNamespaces, NotificationID: defaultNotificationID})
	}

	serverUrl := fmt.Sprintf("%s/notifications/v2?appId=%s&cluster=%s&notifications=%s",
		c.server,
		url.QueryEscape(c.appId),
		url.QueryEscape(c.cluster),
		url.QueryEscape(GetNotification(ns)),
	)

	var res []Notification
	_, err := HttpGetWithTransport(serverUrl, 75, false, &res)

	return res, err
}
