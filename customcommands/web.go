package customcommands

import (
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/common"
	"github.com/jonas747/yagpdb/web"
	"goji.io"
	"goji.io/pat"
	"html/template"
	"net/http"
	"strconv"
	"unicode/utf8"
)

func (p *Plugin) InitWeb() {
	tmplPathSettings := "templates/plugins/customcommands.html"
	if common.Testing {
		tmplPathSettings = "../../customcommands/assets/customcommands.html"
	}

	web.Templates = template.Must(web.Templates.ParseFiles(tmplPathSettings))

	getHandler := web.ControllerHandler(HandleCommands, "cp_custom_commands")

	subMux := goji.SubMux()
	web.CPMux.Handle(pat.New("/customcommands"), subMux)
	web.CPMux.Handle(pat.New("/customcommands/*"), subMux)

	subMux.Use(web.RequireGuildChannelsMiddleware)
	subMux.Use(web.RequireFullGuildMW)

	subMux.Handle(pat.Get(""), getHandler)
	subMux.Handle(pat.Get("/"), getHandler)

	newHandler := web.ControllerPostHandler(HandleNewCommand, getHandler, CustomCommand{}, "Created a new custom command")
	subMux.Handle(pat.Post(""), newHandler)
	subMux.Handle(pat.Post("/"), newHandler)
	subMux.Handle(pat.Post("/:cmd/update"), web.ControllerPostHandler(HandleUpdateCommand, getHandler, CustomCommand{}, "Updated a custom command"))
	subMux.Handle(pat.Post("/:cmd/delete"), web.ControllerHandler(HandleDeleteCommand, "cp_custom_commands"))
}

func HandleCommands(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	client, activeGuild, templateData := web.GetBaseCPContextData(r.Context())

	_, ok := templateData["CustomCommands"]
	if !ok {
		commands, _, err := GetCommands(client, activeGuild.ID)
		if err != nil {
			return templateData, err
		}
		templateData["CustomCommands"] = commands
	}

	return templateData, nil
}

func HandleNewCommand(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	client, activeGuild, templateData := web.GetBaseCPContextData(ctx)
	templateData["VisibleURL"] = "/manage/" + discordgo.StrID(activeGuild.ID) + "/customcommands/"

	newCmd := ctx.Value(common.ContextKeyParsedForm).(*CustomCommand)

	currentCommands, highest, err := GetCommands(client, activeGuild.ID)
	if err != nil {
		return templateData, err
	}

	if len(currentCommands) >= MaxCommands {
		return templateData, web.NewPublicError("Max " + strconv.Itoa(MaxCommands) + " custom commands allowed, if you need more ask on the support server")
	}

	templateData["CustomCommands"] = currentCommands

	newCmd.TriggerType = TriggerTypeFromForm(newCmd.TriggerTypeForm)
	newCmd.ID = highest + 1

	err = newCmd.Save(client, activeGuild.ID)
	if err != nil {
		return templateData, err
	}

	templateData["CustomCommands"] = append(currentCommands, newCmd)
	return templateData, nil
}

func HandleUpdateCommand(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	client, activeGuild, templateData := web.GetBaseCPContextData(ctx)
	templateData["VisibleURL"] = "/manage/" + discordgo.StrID(activeGuild.ID) + "/customcommands/"

	cmd := ctx.Value(common.ContextKeyParsedForm).(*CustomCommand)

	// Validate that they haven't messed with the id
	exists, _ := common.RedisBool(client.Cmd("HEXISTS", KeyCommands(activeGuild.ID), cmd.ID))
	if !exists {
		return templateData, web.NewPublicError("That command dosen't exist?")
	}

	cmd.TriggerType = TriggerTypeFromForm(cmd.TriggerTypeForm)

	err := cmd.Save(client, activeGuild.ID)

	return templateData, err
}

func HandleDeleteCommand(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	ctx := r.Context()
	client, activeGuild, templateData := web.GetBaseCPContextData(ctx)
	templateData["VisibleURL"] = "/manage/" + discordgo.StrID(activeGuild.ID) + "/customcommands/"

	cmdIndex := pat.Param(r, "cmd")

	err := client.Cmd("HDEL", KeyCommands(activeGuild.ID), cmdIndex).Err
	if err != nil {
		return templateData, err
	}

	user := ctx.Value(common.ContextKeyUser).(*discordgo.User)
	go common.AddCPLogEntry(user, activeGuild.ID, "Deleted command #"+cmdIndex)

	return HandleCommands(w, r)
}

func TriggerTypeFromForm(str string) CommandTriggerType {
	switch str {
	case "prefix":
		return CommandTriggerStartsWith
	case "regex":
		return CommandTriggerRegex
	case "contains":
		return CommandTriggerContains
	case "exact":
		return CommandTriggerExact
	default:
		return CommandTriggerCommand

	}
}

func CheckLimits(in ...string) bool {
	for _, v := range in {
		if utf8.RuneCountInString(v) > 2000 {
			return false
		}
	}
	return true
}
