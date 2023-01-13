package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog"
	"strconv"
	"strings"
	"sync"
)

const (
	slashCmdRegisterAddress  = "register-address"
	slashCmdRegisterLocation = "register-location"
	slashCmdRemoveLocation   = "remove-location"
	slashCmdListLocations    = "list-locations"
)

type (
	pwDiscordBot struct {
		botId    string
		session  *discordgo.Session
		user     *discordgo.User
		commands map[string]*discordgo.ApplicationCommand

		userDms sync.Map
		log     zerolog.Logger
	}
)

func (b *pwDiscordBot) setup(token string) error {
	var err error
	b.session, err = discordgo.New("Bot " + token)
	if nil != err {
		return fmt.Errorf("failed to log into discord: %s", err)
	}

	b.user, err = b.session.User("@me")
	if nil != err {
		return fmt.Errorf("failed to get my own user information: %s", err)
	}
	b.botId = b.user.ID
	b.log.Info().Str("User", b.user.String()).Msg("Bot User")

	b.session.AddHandler(b.handleMessageCreate)
	//MessageReactionAdd, MessageReactionRemove, MessageUpdate, MessageDelete
	// MessageAck
	return nil
}

func (b *pwDiscordBot) stop() error {
	return b.session.Close()
}

func (b *pwDiscordBot) RegisterCommandsForGuild(guild *discordgo.UserGuild, nukeExisting bool) {
	b.log.Info().
		Str("Guild", guild.Name).
		Str("GuildID", guild.ID).
		Msgf("Registering slash commands for %s", guild.Name)
	b.commands[slashCmdRegisterAddress] = &discordgo.ApplicationCommand{
		Name:        slashCmdRegisterAddress,
		Description: "We use the address you give to do a lookup to get your geo location",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "A name for you to know what this place is",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "address",
				Description: "The address of the place you want to register",
				Required:    true,
			},
		},
	}

	b.commands[slashCmdRegisterLocation] = &discordgo.ApplicationCommand{
		Name:        slashCmdRegisterLocation,
		Description: "Add a known Latitude/Longitude for your alert",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "A name for you to know what this place is",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "latitude",
				Description: "The latitude component",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "longitude",
				Description: "The longitude component",
				Required:    true,
			},
		},
	}

	b.commands[slashCmdRemoveLocation] = &discordgo.ApplicationCommand{
		Name:        slashCmdRemoveLocation,
		Description: "Removes a location given its name",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "name",
				Description:  "A name you gave this location",
				Required:     true,
				Autocomplete: true,
			},
		},
	}

	b.commands[slashCmdListLocations] = &discordgo.ApplicationCommand{
		Name:        slashCmdListLocations,
		Description: "Lists all your alert locations",
	}

	existingCommands, err := b.session.ApplicationCommands(b.session.State.User.ID, guild.ID)
	if nil != err {
		b.log.Info().Msgf("Unable to list existing slash commands: %s", err)
	}
	if nukeExisting {
		b.log.Info().Msgf("Nuking Existing Commands")
		for _, existingCommand := range existingCommands {
			if err = b.session.ApplicationCommandDelete(b.session.State.User.ID, guild.ID, existingCommand.ID); nil != err {
				b.log.Info().Msgf("Failed to remove command %s - %s", existingCommand.Name, err)
			}
		}
		existingCommands = []*discordgo.ApplicationCommand{}
	}

	for _, cmd := range b.commands {
		exists := false
		for _, existingCommand := range existingCommands {
			if existingCommand.Name == cmd.Name {
				//b.session.ApplicationCommandDelete(b.session.State.User.ID, guild.ID, existingCommand.ID)
				b.log.Info().Msgf("[%s] Command exists: %s - no need to create", guild.Name, cmd.Name)
				exists = true
				break
			}
		}
		if !exists {
			_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, guild.ID, cmd)
			if nil != err {
				b.log.Info().Msgf("[%s] Failed to register command %s: %s", guild.Name, cmd.Name, err)
			} else {
				b.log.Info().Msgf("[%s] Registered Command: %s", guild.Name, cmd.Name)
			}
		}
	}

	// todo: remove commands that we no longer want

	// now register our slash command listener
	b.session.AddHandler(b.handleSlashCommands)
}

func (b *pwDiscordBot) RegisterCommands(nukeExisting bool) {
	for _, guild := range b.GuildList() {
		b.RegisterCommandsForGuild(guild, nukeExisting)
	}
}

// GuildList returns the list of guilds the current user is associated with (handling the error from the API)
func (b *pwDiscordBot) GuildList() []*discordgo.UserGuild {
	list, err := b.session.UserGuilds(200, "", "")
	if nil != err {
		b.log.Info().Msgf("Failed to get list of user guilds for bot: %s", err)
		return []*discordgo.UserGuild{}
	}
	return list
}

// handleMessageCreate is our event handler for when the bot gets a message from a user
func (b *pwDiscordBot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == b.botId {
		return
	}
	isPm := m.GuildID == "" && m.Member == nil
	if !isPm {
		return
	}

	switch strings.ToLower(m.Content) {
	case "list":
		b.sendUserLocationList(m.Author.ID)
	case "hi", "hai", "omg", "omghai", "yo", "sup", "wassup":
		_, err := s.ChannelMessageSend(m.ChannelID, "Hello there.")
		if nil != err {
			b.log.Info().Msgf("Failed to send message. %s", err)
		}
	default:
		var buf strings.Builder
		buf.WriteString("Hello, I am Birdy :small_airplane: the Plane.Watch bot\n")
		buf.WriteString("I can let you know when a plane fly's over your house (or whatever).\n")
		buf.WriteString("You can use any of my commands to interact with me\n")
		buf.WriteString("```\n")
		for _, cmd := range b.commands {
			buf.WriteString(fmt.Sprintf("  * /%-17s - %s\n", cmd.Name, cmd.Description))
		}
		buf.WriteString("```")
		buf.WriteString("`list` will also show you the list of configured alert locations, in case you don't want to use the command\n")

		_, err := s.ChannelMessageSend(m.ChannelID, buf.String())
		if nil != err {
			b.log.Info().Msgf("Failed to send message. %s", err)
		}
	}
}

func (b *pwDiscordBot) handleSlashCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var discordUserName, discordUserId string
	if nil != i.User {
		discordUserId = i.User.ID
		discordUserName = i.User.Username + "#" + i.User.Discriminator
	} else if nil != i.Member {
		discordUserId = i.Member.User.ID
		discordUserName = i.Member.User.Username + "#" + i.Member.User.Discriminator
	} else {
		b.log.Info().Msgf("Failed to determine who I am interacting with")
		return
	}

	userChan, err := s.UserChannelCreate(discordUserId)
	if nil != err {
		b.log.Info().Msgf("Failed to create DM channel for user: %s", err)
	}
	sendPm := func(msg string) {
		if nil == userChan {
			b.log.Info().Msgf("No User Channel - would have said: %s", msg)
			return
		}
		_, err = s.ChannelMessageSend(userChan.ID, msg)
	}
	respondToInteraction := func(msg string) {
		if nil == i {
			return
		}
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Flags:   1 << 6,
			},
		})
		if nil != err {
			b.log.Info().Msgf("Failed to respond to interaction. %s", err)
		}

	}

	switch i.ApplicationCommandData().Name {
	case slashCmdRegisterAddress:
		name := i.ApplicationCommandData().Options[0].StringValue()
		addr := i.ApplicationCommandData().Options[1].StringValue()
		b.log.Info().
			Str("User", discordUserName).
			Str("Location", name).
			Str("Address", addr).
			Msg("User is registering address")
		sendPm(fmt.Sprintf("You are registering address for `%s` ```%s```", name, addr))
		lat, lon, err := geoCodeAddress(addr)
		if nil != err {
			sendPm(err.Error())
			b.log.Info().Msgf("Geolocation failed %s", err)
			respondToInteraction("Geolocation failed, See PM")
		} else {
			if err = addAlertLocation(discordUserId, discordUserName, name, lat, lon); nil != err {
				respondToInteraction("Saving Failed, See PM")
				sendPm(fmt.Sprintf("There was an error adding your address. %s", err))
			} else {
				sendPm("We have setup your alert")
				if err = setLocationAddress(discordUserId, name, addr); nil != err {
					b.log.Info().Msgf("Failed to update user address: %s", err)
					respondToInteraction("Saving Partially Failed, See PM")
					sendPm("Failed to set alert locations address")
				} else {
					respondToInteraction("Success!, More details in PM")
				}
			}
		}
	case slashCmdRegisterLocation:
		name := i.ApplicationCommandData().Options[0].StringValue()
		latS := i.ApplicationCommandData().Options[1].StringValue()
		lonS := i.ApplicationCommandData().Options[2].StringValue()
		lat, errLatConv := strconv.ParseFloat(latS, 64)
		lon, errLonConv := strconv.ParseFloat(lonS, 64)
		if nil != errLatConv {
			sendPm("Was not able to convert your latitude into a decimal latitude")
			respondToInteraction("Failed, More info in PM")
			return
		}
		if nil != errLonConv {
			sendPm("Was not able to convert your longitude into a decimal latitude")
			respondToInteraction("Failed, More info in PM")
			return
		}

		b.log.Info().
			Str("User", discordUserName).
			Str("Location", name).
			Floats64("Lat/Lon", []float64{lat, lon}).
			Msg("user is registering location")
		sendPm(fmt.Sprintf("Adding Location `%s` ```Lat: %0.5f, Lon: %0.5f```", name, lat, lon))

		if err = addAlertLocation(discordUserId, discordUserName, name, lat, lon); nil != err {
			sendPm(fmt.Sprintf("Failed adding that location. ```%s```", err))
			respondToInteraction("Failed, More info in PM")
		} else {
			sendPm("Successfully added alert location")
			respondToInteraction("Success!, More details in PM")
		}

	case slashCmdRemoveLocation:
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			name := i.ApplicationCommandData().Options[0].StringValue()

			b.log.Info().
				Str("User", discordUserName).
				Str("Location", name).
				Msg("user is removing location")
			sendPm(fmt.Sprintf("Removing Location `%s`", name))

			if err = removeAlertLocation(discordUserId, name); nil != err {
				sendPm(fmt.Sprintf("Failed removing location. ```%s```", err))
				respondToInteraction("Failed, see PM")
			} else {
				sendPm("Successfully removed alert location")
				respondToInteraction("Location Removed")
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			var choices []*discordgo.ApplicationCommandOptionChoice
			for _, loc := range getLocationsForUser(discordUserId) {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  loc.LocationName,
					Value: loc.LocationName,
				})
			}
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{Choices: choices},
			})
			if err != nil {
				b.log.Error().Err(err).Msgf("Failed to respond to interaction.")
			}
		}
	case slashCmdListLocations:
		b.sendUserLocationList(discordUserId)
		respondToInteraction("See PM For List")
	default:
		respondToInteraction("Unknown Command")
	}
}

func (b *pwDiscordBot) userDmChan(discordUserId string) (string, error) {
	if _, ok := b.userDms.Load(discordUserId); !ok {
		// need to create a user channel
		userChan, err := b.session.UserChannelCreate(discordUserId)
		if nil != err {
			b.log.Info().Err(err).Msg("Failed to create DM channel for user")
			return "", err
		}
		b.userDms.Store(discordUserId, userChan)
	}

	userChanInterface, ok := b.userDms.Load(discordUserId)
	if !ok {
		return "", errors.New("failed to get the user DM channel from the map")
	}

	userChan, ok := userChanInterface.(*discordgo.Channel)
	if !ok {
		b.userDms.Delete(discordUserId)
		return "", errors.New("what I got from the map was not the right thing")
	}
	return userChan.ID, nil
}

func (b *pwDiscordBot) sendDirectMessage(discordUserId, message string) {
	chanId, err := b.userDmChan(discordUserId)
	if nil != err {
		b.log.Error().Err(err).Msgf("Would have told user [%s] %s", discordUserId, message)
		return
	}

	_, err = b.session.ChannelMessageSend(chanId, message)
	if nil != err {
		b.log.Info().Msgf("Failed to send message: %s", err)
		b.log.Info().Msgf("Would have told user [%s] %s", discordUserId, message)
		// if we are failing on sending, let's just nuke what we have and start again
		b.userDms.Delete(discordUserId)
	}
}

func (b *pwDiscordBot) sendDirectEmbedMsg(discordUserId string, embed *discordgo.MessageEmbed) {
	chanId, err := b.userDmChan(discordUserId)
	if nil != err {
		b.log.Error().Err(err).Msgf("Would have told user [%s] %s", discordUserId, embed.Title)
		return
	}

	_, err = b.session.ChannelMessageSendEmbed(chanId, embed)
	if nil != err {
		b.log.Error().Err(err).Msg("Failed to send message embed")
		b.log.Error().Err(err).Msgf("Would have told user [%s] %s", discordUserId, embed.Title)
		// if we are failing on sending, let's just nuke what we have and start again
		b.userDms.Delete(discordUserId)
	}
}

func (b *pwDiscordBot) sendUserLocationList(discordUserId string) {
	var buf strings.Builder
	buf.WriteString("*Configured Alert Locations*\n")
	buf.WriteString("```\n")
	tbl := tablewriter.NewWriter(&buf)
	tbl.SetColWidth(100)
	tbl.SetHeader([]string{"Name", "Address", "Lat", "Lon"})
	for _, loc := range getLocationsForUser(discordUserId) {
		tbl.Append([]string{
			loc.LocationName,
			loc.Address,
			fmt.Sprintf("%0.5f", loc.Lat),
			fmt.Sprintf("%0.5f", loc.Lon),
		})
	}

	tbl.Render()
	buf.WriteString("```")
	b.sendDirectMessage(discordUserId, buf.String())
}

func (b *pwDiscordBot) sendPlaneAlert(pa *proximityAlert) {
	if nil == pa || nil == pa.alert || nil == pa.update {
		return
	}
	b.log.
		Debug().
		Str("User", pa.alert.DiscordUserName).
		Uint32("Plane", pa.update.Icao).
		Int("Distance (m)", pa.distanceMtr).
		Msg("Alerting user of plane")

	e := discordgo.MessageEmbed{
		URL:         "", // todo: set to plane.watch URL
		Title:       fmt.Sprintf("Proximity Alert for %s", pa.alert.LocationName),
		Description: fmt.Sprintf("%s has entered your airspace", pa.update.Plane()),
		//Timestamp:   "",
		//Color:       0,
		Footer:    nil,
		Image:     nil,
		Thumbnail: nil,
		Video:     nil,
		Provider:  nil,
		Author:    nil,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Aircraft Type",
				Value:  pa.update.AirframeType.Describe(),
				Inline: false,
			},
		},
	}

	b.sendDirectEmbedMsg(pa.alert.DiscordUserId, &e)
}
