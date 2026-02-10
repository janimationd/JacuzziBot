package tamas

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/db"
)

const minigameRoleName string = "Tama Pet Owners"

func AddUserToMinigameRole(session *discordgo.Session, serverId string, userId string) error {
	roleId := db.GetTamaMinigameRole(serverId)
	if roleId == "" {
		mentionable := true
		role, err := session.GuildRoleCreate(serverId, &discordgo.RoleParams{
			Name:        minigameRoleName,
			Mentionable: &mentionable,
		})
		if err != nil {
			log.Println("Failed to create minigame role:", err)
			return err
		} else {
			log.Println("Created minigame role.")
		}
		roleId = role.ID
		err = db.RegisterTamaMinigameRole(serverId, roleId)
		if err != nil {
			log.Println("Failed to register newly created minigame role:", err)
			// We opt to continue rather than give up, this situation sucks regardless.
		}
	}
	return session.GuildMemberRoleAdd(serverId, userId, roleId)
}
