package tickets

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"strconv"
)

type ReopenCommand struct {
}

func (c ReopenCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "reopen",
		Description:     i18n.HelpRename, // TODO: Change
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permission.Everyone,
		Category:        command.Tickets,
		Arguments: command.Arguments(
			command.NewRequiredAutocompleteableArgument("ticket_id", "ID of the ticket to reopen", interaction.OptionTypeInteger, i18n.MessageInvalidArgument, c.AutoCompleteHandler),
		),
	}
}

func (c ReopenCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ReopenCommand) Execute(ctx registry.CommandContext, ticketId int) {
	logic.ReopenTicket(ctx, ticketId)
}

func (ReopenCommand) AutoCompleteHandler(data interaction.ApplicationCommandAutoCompleteInteraction, value string) []interaction.ApplicationCommandOptionChoice {
	if data.GuildId.Value == 0 {
		return nil
	}

	if data.Member == nil {
		return nil
	}

	tickets, err := dbclient.Client.Tickets.GetClosedByUserPrefixed(data.GuildId.Value, data.Member.User.Id, value, 25)
	if err != nil {
		sentry.Error(err)
		return nil
	}

	choices := make([]interaction.ApplicationCommandOptionChoice, len(tickets))
	for i, ticket := range tickets {
		if i >= 25 { // Infallible
			break
		}

		choices[i] = interaction.ApplicationCommandOptionChoice{
			Name:  strconv.Itoa(ticket.Id),
			Value: ticket.Id,
		}
	}

	return choices
}
