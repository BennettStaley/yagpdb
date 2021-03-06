package stdcommands

import (
	"github.com/jonas747/yagpdb/bot/eventsystem"
	"github.com/jonas747/yagpdb/commands"
	"github.com/jonas747/yagpdb/common"
	"github.com/jonas747/yagpdb/stdcommands/advice"
	"github.com/jonas747/yagpdb/stdcommands/allocstat"
	"github.com/jonas747/yagpdb/stdcommands/banserver"
	"github.com/jonas747/yagpdb/stdcommands/calc"
	"github.com/jonas747/yagpdb/stdcommands/catfact"
	"github.com/jonas747/yagpdb/stdcommands/currentshard"
	"github.com/jonas747/yagpdb/stdcommands/currenttime"
	"github.com/jonas747/yagpdb/stdcommands/customembed"
	"github.com/jonas747/yagpdb/stdcommands/define"
	"github.com/jonas747/yagpdb/stdcommands/info"
	"github.com/jonas747/yagpdb/stdcommands/invite"
	"github.com/jonas747/yagpdb/stdcommands/leaveserver"
	"github.com/jonas747/yagpdb/stdcommands/listroles"
	"github.com/jonas747/yagpdb/stdcommands/memberfetcher"
	"github.com/jonas747/yagpdb/stdcommands/mentionrole"
	"github.com/jonas747/yagpdb/stdcommands/ping"
	"github.com/jonas747/yagpdb/stdcommands/poll"
	"github.com/jonas747/yagpdb/stdcommands/reverse"
	"github.com/jonas747/yagpdb/stdcommands/roll"
	"github.com/jonas747/yagpdb/stdcommands/secretcommand"
	"github.com/jonas747/yagpdb/stdcommands/stateinfo"
	"github.com/jonas747/yagpdb/stdcommands/throw"
	"github.com/jonas747/yagpdb/stdcommands/topcommands"
	"github.com/jonas747/yagpdb/stdcommands/topevents"
	"github.com/jonas747/yagpdb/stdcommands/topic"
	"github.com/jonas747/yagpdb/stdcommands/topservers"
	"github.com/jonas747/yagpdb/stdcommands/unbanserver"
	"github.com/jonas747/yagpdb/stdcommands/weather"
	"github.com/jonas747/yagpdb/stdcommands/wouldyourather"
	"github.com/jonas747/yagpdb/stdcommands/yagstatus"
)

type Plugin struct{}

func (p *Plugin) InitBot() {
	commands.AddRootCommands(
		// Info
		info.Command,
		invite.Command,

		// Standard
		define.Command,
		reverse.Command,
		weather.Command,
		calc.Command,
		topic.Command,
		catfact.Command,
		advice.Command,
		ping.Command,
		throw.Command,
		roll.Command,
		customembed.Command,
		currenttime.Command,
		mentionrole.Command,
		listroles.Command,
		wouldyourather.Command,
		poll.Command,

		// Maintenance
		stateinfo.Command,
		secretcommand.Command,
		leaveserver.Command,
		banserver.Command,
		allocstat.Command,
		unbanserver.Command,
		topservers.Command,
		topcommands.Command,
		topevents.Command,
		currentshard.Command,
		memberfetcher.Command,
		yagstatus.Command,
	)
	eventsystem.AddHandler(ping.HandleMessageCreate, eventsystem.EventMessageCreate)
}

func (p *Plugin) Name() string {
	return "stdcommands"
}

func RegisterPlugin() {
	common.RegisterPlugin(&Plugin{})
}
