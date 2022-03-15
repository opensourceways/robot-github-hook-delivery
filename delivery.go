package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/opensourceways/community-robot-lib/githubclient"
	"github.com/opensourceways/community-robot-lib/mq"
	"github.com/sirupsen/logrus"
)

type delivery struct {
	wg   sync.WaitGroup
	hmac func() []byte
}

func (c *delivery) wait() {
	c.wg.Wait()
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func (c *delivery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventType, guid, payload, ok, _ := githubclient.ValidateWebhook(w, r, c.hmac)
	if !ok {
		return
	}

	fmt.Fprint(w, "Event received. Have a nice day.")

	l := logrus.WithFields(
		logrus.Fields{
			"event-type": eventType,
			"event-id":   guid,
		},
	)

	if err := c.publish(payload, r.Header, l); err != nil {
		l.WithError(err).Error()
	}
}

func (c *delivery) publish(payload []byte, h http.Header, l *logrus.Entry) error {
	header := map[string]string{
		"content-type":      h.Get("content-type"),
		"X-GitHub-Event":     h.Get("X-GitHub-Event"),
		"X-GitHub-Delivery": h.Get("X-GitHub-Delivery"),
		"X-Hub-Signature":     h.Get("X-Hub-Signature"),
		"User-Agent":        "Robot-Github-Access",
	}

	msg := mq.Message{
		Header: header,
		Body:   payload,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		if err := mq.Publish("github-webhook", &msg); err != nil {
			l.WithError(err).Error("failed to publish msg")
		} else {
			l.Info("publish message to github-webhook topic success")
		}
	}()

	return nil
}
