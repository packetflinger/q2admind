package main

import (
	"crypto/md5"
	"fmt"
	"strings"
)

/**
 * Each player on a game server has one of these.
 * Each game server has a slice of all current players
 */
type Player struct {
	clientid         int // ID on the gameserver (0-maxplayers)
	database_id      int64
	name             string
	userinfo         string
	userinfomap      map[string]string
	hash             string
	frags            int
	deaths           int
	suicides         int
	teleports        int
	lastteleport     int64 // actually going
	lastteleportlist int64 // viewing the big list of destinations
	invites          int
	lastinvite       int64
	invitesavailable int
	ip               string
	port             int
	fov              int
	connecttime      int64
}

/**
 * Get a pointer to a player based on a client number
 */
func (srv *Server) FindPlayer(cl int) *Player {
	if !srv.ValidPlayerID(cl) {
		return nil
	}

	p := &srv.players[cl]

	if p.connecttime > 0 {
		return p
	}

	return nil
}

/**
 * A player hash is a way of uniquely identifiying a player.
 *
 * It's the first 16 characters of an MD5 hash of their
 * name + skin + fov + partial IP. The idea is to identify
 * players with the same name as different people, so someone can't
 * impersonate someone else and tank their stats.
 *
 * Players can specify a player hash in their Userinfo rather than
 * having one generated. This way they can use different names and
 * still have their stats follow them.
 *
 * To specify a player hash from your q2 config:
 * set phash "<hash here>" u
 */
func LoadPlayerHash(player *Player) {
	var database_id int64

	phash := player.userinfomap["phash"]
	if phash != "" {
		player.hash = phash
	} else {
		ipslice := strings.Split(player.ip, ".")
		ip := fmt.Sprintf("%s.%s.%s", ipslice[0], ipslice[1], ipslice[2])

		pt := []byte(fmt.Sprintf(
			"%s-%s-%s-%s",
			player.name,
			player.userinfomap["skin"],
			player.userinfomap["fov"],
			ip,
		))

		hash := md5.Sum(pt)
		player.hash = fmt.Sprintf("%x", hash[:8])
	}

	database_id = int64(GetPlayerIdFromHash(player.hash))
	if database_id > 0 {
		player.database_id = database_id
		return
	}

	database_id = InsertPlayer(player)
	player.database_id = database_id
}

/**
 * Check if a client ID is valid for a particular server context,
 * does not care if a valid player structure is located there or not
 */
func (srv *Server) ValidPlayerID(cl int) bool {
	return cl >= 0 && cl < len(srv.players)
}

/**
 * Remove a player from the players slice (used when player quits)
 */
func (srv *Server) RemovePlayer(cl int) {
	if srv.ValidPlayerID(cl) {
		srv.players[cl] = Player{}
	}
}

/**
 * Send a message to every player on the server
 */
func (srv *Server) SayEveryone(level int, text string) {
	WriteByte(SCMDSayAll, &srv.messageout)
	WriteByte(byte(level), &srv.messageout)
	WriteString(text, &srv.messageout)
}

/**
 * Send a message to a particular player
 */
func (srv *Server) SayPlayer(client int, level int, text string) {
	WriteByte(SCMDSayClient, &srv.messageout)
	WriteByte(byte(client), &srv.messageout)
	WriteByte(byte(level), &srv.messageout)
	WriteString(text, &srv.messageout)
}

/**
 * Take a back-slash delimited string of userinfo and return
 * a key/value map
 */
func UserinfoMap(ui string) map[string]string {
	info := make(map[string]string)
	if ui == "" {
		return info
	}

	data := strings.Split(ui[1:], "\\")

	for i := 0; i < len(data); i += 2 {
		info[data[i]] = data[i+1]
	}

	// special case: split the IP value into IP and Port
	ip := info["ip"]
	ipport := strings.Split(ip, ":")
	if len(ipport) >= 2 {
		info["port"] = ipport[1]
		info["ip"] = ipport[0]
	}

	return info
}
