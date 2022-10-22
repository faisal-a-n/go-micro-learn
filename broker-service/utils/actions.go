package utils

var actions = map[string]bool{
	"authenticate": true,
}

func IsSupportedAction(action string) bool {
	return actions[action]
}
