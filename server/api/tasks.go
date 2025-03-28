package api

import (
	"github.com/UnifyEM/UnifyEM/common/fields"
	"github.com/UnifyEM/UnifyEM/server/queue"
)

// ProcessMessageQueue reads queued messages from agents
// This function is called periodically from tasks in main.go
// It is in this package because it is processing messages that
// arrived from the API and requires access to the underlying data
// layer to store the messages in the agent metadata
func (a *API) ProcessMessageQueue() {
	for {
		message, ok := queue.Read()
		if !ok {
			break
		}

		// Log the message
		a.logger.Info(2050, "agent message",
			fields.NewFields(
				fields.NewField("id", message.AgentID),
				fields.NewField("sent", message.Sent),
				fields.NewField("type", message.MessageType),
				fields.NewField("message", message.Message)))

		// Send the message to the data layer
		err := a.data.NewAgentMessage(message)
		if err != nil {
			a.logger.Errorf(2051, "error adding message to event store: %s", err.Error())
		}
	}
}
