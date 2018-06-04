package commands

import (
	"context"
	"database/sql"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/commands/models"
	"github.com/jonas747/yagpdb/common"
	"github.com/jonas747/yagpdb/web"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"github.com/volatiletech/sqlboiler/types"
	"goji.io"
	"goji.io/pat"
	"html/template"
	"net/http"
	"strconv"
)

type ChannelOverrideForm struct {
	Channels                []int64 `valid:"channel,true`
	Global                  bool
	CommandsEnabled         bool
	AutodeleteResponse      bool
	AutodeleteTrigger       bool
	AutodeleteResponseDelay int
	AutodeleteTriggerDelay  int
	RequireRoles            []int64 `valid:"role,true"`
	IgnoreRoles             []int64 `valid:"role,true"`
}

type CommandOverrideForm struct {
}

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

	getHandler := web.ControllerHandler(HandleCommands, "cp_commands")
	subMux.Handle(pat.Get(""), getHandler)
	subMux.Handle(pat.Get("/"), getHandler)
	subMux.Handle(pat.Post("/"), web.ControllerPostHandler(HandlePostCommands, getHandler, nil, "Updated command prefix"))

	// Channel override handlers
	subMux.Handle(pat.Post("/channel_overrides/new"),
		web.ControllerPostHandler(HandleCreateChannelsOverride, getHandler, ChannelOverrideForm{}, "Created a new command channels override"))

	subMux.Handle(pat.Post("/channel_overrides/:channelOverride/update"),
		ChannelOverrideMiddleware(web.ControllerPostHandler(HandleUpdateChannelsOverride, getHandler, ChannelOverrideForm{}, "Updated a commands channel override")))

	subMux.Handle(pat.Post("/channel_overrides/:channelOverride/delete"),
		ChannelOverrideMiddleware(web.ControllerPostHandler(HandleDeleteChannelsOverride, getHandler, nil, "Deleted a commands channel override")))

	// Command override handlers
	subMux.Handle(pat.Post("/channel_overrides/:channelOverride/command_overrides/new"),
		ChannelOverrideMiddleware(web.ControllerPostHandler(HandleCreateCommandOverride, getHandler, CommandOverrideForm{}, "Created a commands command override")))

	subMux.Handle(pat.Post("/channel_overrides/:channelOverride/command_overrides/:commandsOverride/update"),
		ChannelOverrideMiddleware(web.ControllerPostHandler(HandleUpdateCommandOVerride, getHandler, CommandOverrideForm{}, "Updated a commands command override")))

	subMux.Handle(pat.Post("/channel_overrides/:channelOverride/command_overrides/:commandsOverride/delete"),
		ChannelOverrideMiddleware(web.ControllerPostHandler(HandleDeleteCommandOverride, getHandler, nil, "Deleted a commands command override")))

}

// Servers the command page with current config
func HandleCommands(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	client, activeGuild, templateData := web.GetBaseCPContextData(ctx)

	type SortedCommands struct {
		Category string
		Commands []string
	}

	commands := make([]*SortedCommands, 0, len(CommandSystem.Root.Commands))
OUTER:
	for _, cmd := range CommandSystem.Root.Commands {
		cast := cmd.Command.(*YAGCommand)
		if cast.HideFromHelp {
			continue
		}

		for _, v := range commands {
			if v.Category == cast.CmdCategory.Name {
				v.Commands = append(v.Commands, cmd.Trigger.Names[0])
				continue OUTER
			}
		}

		commands = append(commands, &SortedCommands{
			Category: cast.CmdCategory.Name,
			Commands: []string{cmd.Trigger.Names[0]},
		})
	}

	templateData["SortedCommands"] = commands

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

	prefix, _ := GetCommandPrefix(client, activeGuild.ID)

	templateData["CommandPrefix"] = prefix

	templateData["VisibleURL"] = "/manage/" + discordgo.StrID(activeGuild.ID) + "/commands/settings"

	return templateData, nil
}

// Handles the updating of global and per channel command settings
func HandlePostCommands(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	client, activeGuild, templateData := web.GetBaseCPContextData(ctx)
	newPrefix := r.FormValue("Prefix")
	if len(newPrefix) < 1 || len(newPrefix) > 100 {
		return templateData, web.NewPublicError("Prefix is smaller than 1 or larger than 100 characters")
	}

	client.Cmd("SET", "command_prefix:"+discordgo.StrID(activeGuild.ID), newPrefix)

	return templateData, nil
}

// Channel override handlers
func ChannelOverrideMiddleware(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		activeGuild := r.Context().Value(common.ContextKeyCurrentGuild).(*discordgo.Guild)

		var override *models.CommandsChannelsOverride
		var err error

		id := pat.Param(r, "channelOverride")
		if id == "global" {
			override, err = models.CommandsChannelsOverridesG(qm.Where("guild_id = ? AND global=true", activeGuild.ID)).One()
			if err == sql.ErrNoRows {
				override = &models.CommandsChannelsOverride{
					Global: true,
				}

				err = nil
			}
		} else {
			idParsed, _ := strconv.ParseInt(id, 10, 64)
			override, err = models.CommandsChannelsOverridesG(qm.Where("guild_id = ? AND id = ?", activeGuild.ID, idParsed)).One()
		}

		if err != nil {
			_, tmpl := web.GetCreateTemplateData(r.Context())
			tmpl.AddAlerts(web.ErrorAlert("Channels override not found"))
			inner.ServeHTTP(w, r)
			return
		}

		inner.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyChannelOverride, override)))
	})
}

func HandleCreateChannelsOverride(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	_, activeGuild, templateData := web.GetBaseCPContextData(r.Context())
	formData := r.Context().Value(common.ContextKeyParsedForm).(*ChannelOverrideForm)

	count, err := models.CommandsChannelsOverridesG(qm.Where("guild_id = ?", activeGuild.ID), qm.Where("channels && ?", types.Int64Array(formData.Channels))).Count()
	if err != nil {
		return templateData, errors.WithMessage(err, "count")
	}

	if count > 0 {
		return templateData.AddAlerts(web.ErrorAlert("One of the selected channels is already used in another override")), nil
	}

	model := &models.CommandsChannelsOverride{
		GuildID:                 activeGuild.ID,
		Channels:                formData.Channels,
		Global:                  formData.Global,
		CommandsEnabled:         formData.CommandsEnabled,
		AutodeleteResponse:      formData.AutodeleteResponse,
		AutodeleteTrigger:       formData.AutodeleteTrigger,
		AutodeleteResponseDelay: formData.AutodeleteResponseDelay,
		AutodeleteTriggerDelay:  formData.AutodeleteTriggerDelay,
		RequireRoles:            formData.RequireRoles,
		IgnoreRoles:             formData.IgnoreRoles,
	}

	err = model.InsertG()
	return templateData, errors.WithMessage(err, "InsertG")
}

func HandleUpdateChannelsOverride(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	_, activeGuild, templateData := web.GetBaseCPContextData(r.Context())

	currentOverrideInterface := r.Context().Value(CtxKeyChannelOverride)
	if currentOverrideInterface == nil {
		return templateData, nil
	}

	currentOverride := currentOverrideInterface.(*models.CommandsChannelsOverride)

	formData := r.Context().Value(common.ContextKeyParsedForm).(*ChannelOverrideForm)

	count, err := models.CommandsChannelsOverridesG(
		qm.Where("guild_id = ?", activeGuild.ID), qm.Where("channels && ?", types.Int64Array(formData.Channels)), qm.Where("id != ?", currentOverride.ID)).Count()

	if err != nil {
		return templateData, errors.WithMessage(err, "count")
	}

	if count > 0 {
		return templateData.AddAlerts(web.ErrorAlert("One of the selected channels is already used in another override")), nil
	}

	currentOverride.Channels = formData.Channels
	currentOverride.Global = formData.Global
	currentOverride.CommandsEnabled = formData.CommandsEnabled
	currentOverride.AutodeleteResponse = formData.AutodeleteResponse
	currentOverride.AutodeleteTrigger = formData.AutodeleteTrigger
	currentOverride.AutodeleteResponseDelay = formData.AutodeleteResponseDelay
	currentOverride.AutodeleteTriggerDelay = formData.AutodeleteTriggerDelay
	currentOverride.RequireRoles = formData.RequireRoles
	currentOverride.IgnoreRoles = formData.IgnoreRoles

	err = currentOverride.UpdateG()
	return templateData, errors.WithMessage(err, "UpdateG")
}

func HandleDeleteChannelsOverride(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	_, _, templateData := web.GetBaseCPContextData(r.Context())

	currentOverrideInterface := r.Context().Value(CtxKeyChannelOverride)
	if currentOverrideInterface == nil {
		return templateData, nil
	}

	currentOverride := currentOverrideInterface.(*models.CommandsChannelsOverride)

	err := currentOverride.DeleteG()
	return templateData, errors.WithMessage(err, "DeleteG")
}

// Command handlers
func HandleCreateCommandOverride(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	return nil, nil
}
func HandleUpdateCommandOVerride(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	return nil, nil
}

func HandleDeleteCommandOverride(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	return nil, nil
}
