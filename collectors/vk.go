package collectors

import (
	"fmt"
	"strings"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/Truenya/dota-search-back/condlog"
	"github.com/Truenya/dota-search-back/config"
	"github.com/Truenya/dota-search-back/data"
	"github.com/Truenya/dota-search-back/db"
)

type VKCollector struct {
	Vk *api.VK
}

func NewVKCollector() VKCollector {
	VK := api.NewVK(config.GetToken())

	VK.Limit = 200

	return VKCollector{Vk: VK}
}

func (v *VKCollector) Collect() {
	p := params.NewSearchGetHintsBuilder()

	p.SearchGlobal(true)
	p.Limit(20)
	p.Q("dota")

	groups, err := v.Vk.SearchGetHints(p.Params)
	condlog.PanicCond("%v", err)

	for _, group := range groups.Items {
		if group.Group.ID <= 0 {
			continue
		}

		fmt.Println(group.Group.Name)

		pt := params.NewBoardGetTopicsBuilder()

		pt.Count(100)
		pt.GroupID(group.Group.ID)

		topics, err := v.Vk.BoardGetTopics(pt.Params)
		condlog.PanicCond("[collector] %v", err)

		for _, topic := range topics.Items {
			if strings.Contains(strings.ToLower(topic.Title), "поиск") {
				fmt.Println(group.Group.ID, topic.Title, topic.Comments)
				db.StoreVKConfig(data.GroupData{ID: group.Group.ID, TopicID: topic.ID, MaxMessageID: topic.Comments})
			}
		}
	}
}
