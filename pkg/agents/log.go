package agents

import (
	"encoding/json"
	"fmt"
	"os"
)

type AgentLog struct {
	*os.File

	name string
	msgs []Message
}

func (l *AgentLog) Print(str string) {
	_, _ = l.Write([]byte(str))
}

func (l *AgentLog) Printf(str string, args ...any) {
	_, _ = l.Write([]byte(fmt.Sprintf(str, args...)))
}

func (l *AgentLog) Println(str string) {
	_, _ = l.Write([]byte(str + "\n"))
}

func (l *AgentLog) Message(msg Message) {
	l.Printf("## %s (%s):\n%s\n", msg.From.Role, msg.From.Name, msg.Text)

	_ = l.Sync()

	l.msgs = append(l.msgs, msg)
	data, err := json.Marshal(l.msgs)

	if err == nil {
		_ = os.WriteFile("/tmp/agib-agent-logs/"+l.name+".json", data, 0644)
	}
}

func NewAgentLog(name string) *AgentLog {
	if err := os.MkdirAll("/tmp/agib-agent-logs", 0755); err != nil {
		panic(err)
	}

	f, err := os.OpenFile("/tmp/agib-agent-logs/"+name+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC, 0644)

	if err != nil {
		panic(err)
	}

	return &AgentLog{File: f, name: name}
}
