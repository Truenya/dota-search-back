package miners

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/Truenya/dota-search-back/config"
	"github.com/Truenya/dota-search-back/data"
	"github.com/Truenya/dota-search-back/db"
)

type vk struct {
	Vk       *api.VK
	messages *chan data.Messages
	needStop *chan struct{}
	Groups   []data.GroupData
}

func newVKMiner() vk {
	VK := api.NewVK(config.GetToken())
	VK.Limit = 200
	needStop := make(chan struct{}, 1)

	return vk{
		Vk:       VK,
		needStop: &needStop,
		Groups:   db.GetVKMiningSettings(),
	}
}

func (z vk) Spoof() {
	log.Info("Start mining vk data")

	for _, group := range z.Groups {
		p := params.NewBoardGetCommentsBuilder()
		p.GroupID(group.ID)
		log.Debugln(group.TopicID)
		p.TopicID(group.TopicID)
		p.Sort("desc")
		p.StartCommentID(0)
		p.Count(100)

		z.Vk.Limit = 200
		// get information about the messages
		msgs, err := z.Vk.BoardGetComments(p.Params)
		if err != nil {
			log.Errorln(err)
			continue
		}

		var spoofed data.Messages

		for _, item := range msgs.Items {
			mes := data.Message{
				ID:        item.ID,
				Link:      fmt.Sprintf("https://z.vk.com/im?media=&sel=%v", item.FromID),
				Data:      item.Text,
				Timestamp: time.Unix(int64(item.Date), 0),
			}
			mes.SetAuthorID(item.FromID)
			mes.CalculateHash()
			spoofed.Messages = append(spoofed.Messages, mes)
		}

		*z.messages <- spoofed
	}

	log.Info("Finished mining vk data")
}

func (z vk) Spoofing(messages *chan data.Messages) {
	z.messages = messages
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-*z.needStop:
			return
		case <-ticker.C:
			ticker.Reset(1 * time.Hour)
			z.Spoof()
		}
	}
}

func (z vk) Stop() {
	*z.needStop <- struct{}{}
}
