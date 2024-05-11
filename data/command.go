package data

type Command struct {
	Player
}

type Commands struct {
	Commands []Command `json:"list"`
}

func ToCommands(commands ...Command) Commands {
	return Commands{Commands: commands}
}
