package commands

import (
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/commands/models"
	"github.com/jonas747/yagpdb/common"
	"github.com/jonas747/yagpdb/web"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"goji.io"
	"goji.io/pat"
	"html/template"
	"net/http"
)

func (p *Plugin) InitWeb() {
	tmplPath := "templates/plugins/commands.html"
	if common.Testing {
		tmplPath = "../../commands/assets/commands.html"
	}

	web.Templates = template.Must(web.Templates.ParseFiles(tmplPath))

	subMux := goji.SubMux()
	web.CPMux.Handle(pat.New("/commands/settings"), subMux)
	web.CPMux.Handle(pat.New("/commands/settings/*"), subMux)

	subMux.Use(web.RequireGuildChannelsMiddleware)
	subMux.Use(web.RequireFullGuildMW)

	subMux.Handle(pat.Get(""), web.ControllerHandler(HandleCommands, "cp_commands"))
	subMux.Handle(pat.Get("/"), web.ControllerHandler(HandleCommands, "cp_commands"))
	subMux.Handle(pat.Post("/"), web.RenderHandler(HandlePostCommands, "cp_commands"))
}

// Servers the command page with current config
func HandleCommands(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	client, activeGuild, templateData := web.GetBaseCPContextData(ctx)

	channels := ctx.Value(common.ContextKeyGuildChannels).([]*discordgo.Channel)

	commands := make([]string, 0, len(CommandSystem.Root.Commands))
	for _, cmd := range CommandSystem.Root.Commands {
		commands = append(commands, cmd.Trigger.Names[0])
	}

	templateData["Commands"] = commands

	channelOverrides, err := models.CommandsChannelsOverridesG(qm.Where("guild_id=?", activeGuild.ID), qm.Load("CommandsCommandOverrides")).All()
	if err != nil {
		return templateData, err
	}

	var global *models.CommandsChannelsOverride
	for i, v := range channelOverrides {
		if v.Global {
			global = v
			channelOverrides = append(channelOverrides[:i], channelOverrides[i+1:]...)
			break
		}
	}

	if global == nil {
		global = &models.CommandsChannelsOverride{
			Global: true,
		}
	}

	templateData["GlobalCommandSettings"] = global
	templateData["ChannelOverrides"] = channelOverrides

	templateData["CommandConfig"] = GetConfigLegacy(client, activeGuild.ID, channels)
	return templateData, nil
}

// Handles the updating of global and per channel command settings
func HandlePostCommands(w http.ResponseWriter, r *http.Request) interface{} {
	// ctx := r.Context()
	// client, activeGuild, templateData := web.GetBaseCPContextData(ctx)
	// templateData["VisibleURL"] = "/manage/" + discordgo.StrID(activeGuild.ID) + "/commands/settings"

	return nil
}
