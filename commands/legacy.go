package commands

import (
	"github.com/jonas747/dcmd"
	"github.com/jonas747/discordgo"
	"github.com/jonas747/yagpdb/common"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/sirupsen/logrus"
)

type LegacyChannelCommandSetting struct {
	Info           *YAGCommand `json:"-"` // Used for template info
	Cmd            string      `json:"cmd"`
	CommandEnabled bool        `json:"enabled"`
	AutoDelete     bool        `json:"autodelete"`
	RequiredRole   string      `json:"required_role"`
}

type LegacyChannelOverride struct {
	Settings        []*LegacyChannelCommandSetting `json:"settings"`
	OverrideEnabled bool                           `json:"enabled"`
	Channel         string                         `json:"channel"`
	ChannelName     string                         `json:"-"` // Used for the template rendering
}

type LegacyCommandsConfig struct {
	Prefix string `json:"-"` // Stored in a seperate key for speed

	Global           []*LegacyChannelCommandSetting `json:"gloabl"`
	ChannelOverrides []*LegacyChannelOverride       `json:"overrides"`
}

// Fills in the defaults for missing data, for when users create channels or commands are added
func CheckChannelsConfigLegacy(conf *LegacyCommandsConfig, channels []*discordgo.Channel) {

	commands := CommandSystem.Root.Commands

	if conf.Global == nil {
		conf.Global = []*LegacyChannelCommandSetting{}
	}

	if conf.ChannelOverrides == nil {
		conf.ChannelOverrides = []*LegacyChannelOverride{}
	}

ROOT:
	for _, channel := range channels {
		if channel.Type != discordgo.ChannelTypeGuildText {
			continue
		}
		strCID := discordgo.StrID(channel.ID)
		// Look for an existing override
		for _, override := range conf.ChannelOverrides {
			// Found an existing override, check if it has all the commands
			if strCID == override.Channel {
				override.Settings = checkCommandSettingsLegacy(override.Settings, commands, false)
				override.ChannelName = channel.Name // Update name if changed
				continue ROOT
			}
		}

		// Not found, create a default override
		override := &LegacyChannelOverride{
			Settings:        []*LegacyChannelCommandSetting{},
			OverrideEnabled: false,
			Channel:         strCID,
			ChannelName:     channel.Name,
		}

		// Fill in default command settings
		override.Settings = checkCommandSettingsLegacy(override.Settings, commands, false)
		conf.ChannelOverrides = append(conf.ChannelOverrides, override)
	}

	newOverrides := make([]*LegacyChannelOverride, 0, len(conf.ChannelOverrides))

	// Check for removed channels
	for _, override := range conf.ChannelOverrides {
		for _, channel := range channels {
			if channel.Type != discordgo.ChannelTypeGuildText {
				continue
			}

			if discordgo.StrID(channel.ID) == override.Channel {
				newOverrides = append(newOverrides, override)
				break
			}
		}
	}
	conf.ChannelOverrides = newOverrides

	// Check the global settings
	conf.Global = checkCommandSettingsLegacy(conf.Global, commands, true)
}

// Checks a single list of LegacyChannelCommandSettings and applies defaults if not found
func checkCommandSettingsLegacy(settings []*LegacyChannelCommandSetting, commands []*dcmd.RegisteredCommand, defaultEnabled bool) []*LegacyChannelCommandSetting {

ROOT:
	for _, registeredCmd := range commands {
		cast, ok := registeredCmd.Command.(*YAGCommand)
		if !ok {
			continue
		}

		for _, settingsCmd := range settings {
			if cast.Name == settingsCmd.Cmd {
				// Bingo
				settingsCmd.Info = cast
				continue ROOT
			}
		}

		// Not found, add it to the list of overrides
		settingsCmd := &LegacyChannelCommandSetting{
			Cmd:            cast.Name,
			CommandEnabled: defaultEnabled,
			AutoDelete:     false,
			Info:           cast,
		}
		settings = append(settings, settingsCmd)
	}

	newSettings := make([]*LegacyChannelCommandSetting, 0, len(settings))

	// Check for commands that have been removed (e.g the config contains commands from an older version)
	for _, settingsCmd := range settings {
		for _, registeredCmd := range commands {
			cast, ok := registeredCmd.Command.(*YAGCommand)
			if !ok {
				continue
			}

			if cast.Name == settingsCmd.Cmd {
				newSettings = append(newSettings, settingsCmd)
				break
			}
		}
	}

	return newSettings
}

func GetConfigLegacy(client *redis.Client, guild int64, channels []*discordgo.Channel) *LegacyCommandsConfig {
	var config *LegacyCommandsConfig
	err := common.GetRedisJson(client, "commands_settings:"+discordgo.StrID(guild), &config)
	if err != nil {
		logrus.WithError(err).Error("Error retrieving command settings")
	}

	if config == nil {
		config = &LegacyCommandsConfig{}
	}

	prefix, err := GetCommandPrefix(client, guild)
	if err != nil {
		// Continue as normal with defaults
		logrus.WithError(err).Error("Error fetching command prefix")
	}

	config.Prefix = prefix

	// Fill in defaults
	CheckChannelsConfigLegacy(config, channels)

	return config
}
