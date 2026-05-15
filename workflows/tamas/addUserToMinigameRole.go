package tamas

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/db"
)

const minigameRoleName string = "Tama Pet Owners"

func AddUserToMinigameRole(session *discordgo.Session, serverId string, userId string) error {
	roleId := db.GetTamaMinigameRole(serverId)
	// If no role is in the DB yet.
	if roleId == "" {
		// See if there's already a role with the name we want to use, and reuse it if so.
		existingRoles, err := session.GuildRoles(serverId)
		if err != nil {
			log.Println("Failed to create minigame role:", err)
			return err
		}
		var role *discordgo.Role = nil
		for _, existingRole := range existingRoles {
			if existingRole.Name == minigameRoleName {
				log.Printf("Reusing minigame role %s.\n", existingRole.ID)
				role = existingRole
				break
			}
		}
		// No existing role, create one
		if role == nil {
			mentionable := true
			role, err = session.GuildRoleCreate(serverId, &discordgo.RoleParams{
				Name:        minigameRoleName,
				Mentionable: &mentionable,
			})
			if err != nil {
				log.Printf("Failed to create minigame role: %s\n", err.Error())
				return err
			} else {
				log.Println("Created minigame role.")
			}
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
