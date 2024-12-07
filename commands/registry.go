package commands

type ResponseWriter interface {
    WriteJSON(statusCode int, headers map[string][]string, body interface{}) error
    Stream(eventType string, data interface{}) error
    End() error
}

type Command interface {
    Execute(args map[string]interface{}, writer ResponseWriter) error
}

var registry = make(map[string]Command)

func RegisterCommand(name string, cmd Command) {
    registry[name] = cmd
}

func GetCommand(name string) Command {
    return registry[name]
}
