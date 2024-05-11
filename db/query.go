package db

var queries = map[string]string{
	"deletec":   "delete from command where ip=$1",
	"deletep":   "delete from player where ip=$1",
	"getallc":   "select ip, data, link,mmr, possible_pos, timestamp from command",
	"getalli":   "select ip from command union select ip from player",
	"getallm":   "select message_id, data, link, timestamp, hash from messages",
	"getallu":   "select ip, login, link,mmr, possible_pos, timestamp from player",
	"getvkms":   "select group_id, topic_id, last_message_id from vk_mining_settings",
	"getvkcs":   "select offset_count from vk_collecting_settings",
	"storec":    "insert into command (ip,data,mmr,link,possible_pos,timestamp) values ($1, $2, $3, $4, $5, $6)",
	"storem":    "insert into messages (data, link, timestamp, hash, message_id) values ($1, $2, $3, $4, $5)",
	"storep":    "insert into player (ip,login,mmr,link,possible_pos,timestamp) values ($1, $2, $3, $4, $5, $6)",
	"storevkcs": "update vk_collecting_settings set offset_count = $1 where mining_value='comments'",
	"storevkms": "insert into vk_mining_settings (group_id, topic_id, last_message_id) values ($1, $2, $3)",
	"updatec":   "update command set data=$1,mmr=$2,link=$3,possible_pos=$4,timestamp=$5 where ip = $6",
	"updatep":   "update player set login=$1,mmr=$2,link=$3,possible_pos=$4,timestamp=$5 where ip = $6",
}
