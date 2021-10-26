package manager

import (
	"fmt"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/button"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/command/context"
	cmdregistry "github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/interaction/component"
)

// Returns whether the handler may edit the message
func HandleInteraction(manager *ComponentInteractionManager, worker *worker.Context, data interaction.MessageComponentInteraction, responseCh chan button.Response) bool {
	// Safety checks
	if data.GuildId.Value != 0 && data.Member == nil {
		return false
	}

	if data.GuildId.Value == 0 && data.User == nil {
		return false
	}

	premiumTier, err := getPremiumTier(worker, data)
	if err != nil {
		sentry.ErrorWithContext(err, errorcontext.WorkerErrorContext{
			Guild:   data.GuildId.Value,
			Channel: data.ChannelId,
		})
		return false
	}

	switch data.Data.Type() {
	case component.ComponentButton:
		handler := manager.MatchButton(data.Data.AsButton().CustomId)
		if handler == nil {
			return false
		}

		ctx := context.NewButtonContext(worker, data, premiumTier, responseCh)
		shouldExecute, canEdit := doPropertiesChecks(data, ctx, handler.Properties())
		if shouldExecute {
			go handler.Execute(ctx)
		}

		return canEdit
	case component.ComponentSelectMenu:
		handler := manager.MatchSelect(data.Data.AsSelectMenu().CustomId)
		if handler == nil {
			return false
		}

		ctx := context.NewSelectMenuContext(worker, data, premiumTier, responseCh)
		shouldExecute, canEdit := doPropertiesChecks(data, ctx, handler.Properties())
		if shouldExecute {
			go handler.Execute(ctx)
		}

		return canEdit
	default:
		sentry.ErrorWithContext(fmt.Errorf("invalid message component type: %d", data.Data.ComponentType), errorcontext.WorkerErrorContext{
			Guild:   data.GuildId.Value,
			Channel: data.ChannelId,
		})
		return false
	}
}

func getPremiumTier(worker *worker.Context, data interaction.MessageComponentInteraction) (premium.PremiumTier, error) {
	// Psuedo premium if DM command
	if data.GuildId.Value == 0 {
		if worker.IsWhitelabel {
			return premium.Whitelabel, nil
		} else {
			return premium.Premium, nil
		}
	} else {
		premiumTier, err := utils.PremiumClient.GetTierByGuildId(data.GuildId.Value, true, worker.Token, worker.RateLimiter)
		if err != nil {
			return premium.None, err
		}

		return premiumTier, nil
	}
}

func doPropertiesChecks(data interaction.MessageComponentInteraction, ctx cmdregistry.CommandContext, properties registry.Properties) (shouldExecute, canEdit bool) {
	if data.GuildId.Value == 0 && !properties.HasFlag(registry.DMsAllowed) {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageButtonGuildOnly)
		return false, false
	}

	if data.GuildId.Value != 0 && !properties.HasFlag(registry.GuildAllowed) {
		ctx.Reply(constants.Red, i18n.Error, i18n.MessageButtonDMOnly)
		return false, false
	}

	return true, properties.HasFlag(registry.CanEdit)
}