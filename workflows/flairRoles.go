package workflows

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/janimationd/JacuzziBot/models"
)

func EnsureFlairRoles(session *discordgo.Session, serverId string) error {
	guild, err := session.Guild(serverId)
	if err != nil {
		return fmt.Errorf("Couldn't fetch server roles: %w", err)
	}

	serverRoles := guild.Roles

	for flairLevel, flair := range models.FlairProps {
		// Don't create a role for None
		if models.FlairLevel(flairLevel) == models.FlairNone {
			continue
		}

		roleName := flair.RoleName
		found := false
		for _, role := range serverRoles {
			if role.Name == roleName {
				found = true
				break
			}
		}
		if !found {
			no := false
			_, err := session.GuildRoleCreate(serverId, &discordgo.RoleParams{
				Name:        roleName,
				Color:       &flair.ColorCode,
				Hoist:       &no,
				Mentionable: &no,
			})
			if err != nil {
				return fmt.Errorf("Couldn't create flair role '%s': %w", roleName, err)
			}
		}
	}

	return nil
}

func getServerRoleWithId(roleList []*discordgo.Role, roleId string) *discordgo.Role {
	for _, role := range roleList {
		if role.ID == roleId {
			return role
		}
	}
	panic(fmt.Errorf("Role %s not present in server role list!", roleId))
}

func getServerRoleWithName(roleList []*discordgo.Role, roleName string) *discordgo.Role {
	for _, role := range roleList {
		if role.Name == roleName {
			return role
		}
	}
	panic(fmt.Errorf("Role %s not present in server role list!", roleName))
}

func ChangeUserFlairRole(
	session *discordgo.Session,
	serverId string,
	userId string,
	newFlairLevel models.FlairLevel,
) error {
	member, err := session.GuildMember(serverId, userId)
	if err != nil {
		return fmt.Errorf("Couldn't fetch guild member details when changing user's flair role: %w", err)
	}

	serverRoles, err := session.GuildRoles(serverId)
	if err != nil {
		return fmt.Errorf("Couldn't fetch guild roles when changing user's flair role: %w", err)
	}

	// Unassign them from any other flair roles
	for _, memberRoleId := range member.Roles {
		serverRole := getServerRoleWithId(serverRoles, memberRoleId)
		for flairLevel, flair := range models.FlairProps {
			// Skip flair None (there is no role for it)
			if models.FlairLevel(flairLevel) == models.FlairNone {
				continue
			}

			if flair.RoleName == serverRole.Name {
				err = session.GuildMemberRoleRemove(serverId, userId, memberRoleId)
				if err != nil {
					return fmt.Errorf("Couldn't remove old user flair role: %w", err)
				}
			}
		}
	}

	// Assign them to the new flair role if there is one for their new level
	if newFlairLevel != models.FlairNone {
		newFlair := models.FlairProps[newFlairLevel]
		flairRole := getServerRoleWithName(serverRoles, newFlair.RoleName)

		err = session.GuildMemberRoleAdd(serverId, userId, flairRole.ID)
		if err != nil {
			return fmt.Errorf("Couldn't apply user's new flair role: %w", err)
		}
	}

	return nil
}
