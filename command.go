package main

import (
    "errors"
    "fmt"
    "log"
)

/**
 * Player issued the teleport command.
 *
 * If a destination is supplied, just send the player there,
 * else send a list of possibilities
 */
func Teleport(srv *Server) {
    cl := ReadByte(&srv.message)
    dest := ReadString(&srv.message)
    p := findplayer(srv.players, int(cl))
    log.Printf("[%s/TELEPORT/%s] %s\n", srv.name, p.name, dest)

    if dest == "" {
        txt := "Sorry, teleport command is still under construction\n"
        SayPlayer(srv, int(cl), PRINT_HIGH, txt)
        return
    }

    newserver, err := FindTeleportDestination(dest)
    if err != nil {
        log.Println("warning,", err)
        SayPlayer(srv, int(cl), PRINT_HIGH, "Unknown destination\n")
    } else {
        st := fmt.Sprintf("connect %s\n", newserver)
        StuffPlayer(srv, int(cl), st)
    }

    txt := fmt.Sprintf("TELEPORT [%d] %s", cl, p.name)
    LogEventToDatabase(srv.id, LogTypeCommand, txt)
}

/**
 * Resolve a teleport name to an ip:port
 */
func FindTeleportDestination(dest string) (string, error){
    for _, s := range servers {
        if s.name == dest {
            return fmt.Sprintf("%s:%d", s.ipaddress, s.port), nil
        }
    }

    return "", errors.New("unknown destination")
}

/**
 * Player issued an invite command.
 *
 * Broadcast the invite to all connected servers
 */
func Invite(srv *Server) {
    cl := ReadByte(&srv.message)
    text := ReadString(&srv.message)
    p := findplayer(srv.players, int(cl))
    log.Printf("[%s/INVITE/%s] %s\n", srv.name, p.name, text)

    //txt := "Sorry, INVITE command is currently under construction\n"
    //SayPlayer(srv, int(cl), PRINT_HIGH, txt)
    //StuffPlayer(srv, int(cl), "say this better work")
    MutePlayer(srv, p.clientid, 15)
}

/**
 * Force a player to do a command
 */
func StuffPlayer(srv *Server, cl int, cmd string) {
    stuffcmd := fmt.Sprintf("sv !stuff CL %d %s\n", cl, cmd)
    WriteByte(SCMDCommand, &srv.messageout)
    WriteString(stuffcmd, &srv.messageout)
}

/**
 * Temporarily prevent the player from talking
 * using a negative number of seconds makes it
 * permanent.
 */
func MutePlayer(srv *Server, cl int, seconds int) {
    cmd := ""
    if seconds < 0 {
        cmd = fmt.Sprintf("sv !mute CL %d PERM\n", cl)
    } else {
        cmd = fmt.Sprintf("sv !mute CL %d %d", cl, seconds)
    }
    WriteByte(SCMDCommand, &srv.messageout)
    WriteString(cmd, &srv.messageout)

    txt := fmt.Sprintf("MUTE [%d] was muted")
    LogEventToDatabase(srv.id, LogTypeCommand, txt)
}

/**
 *
 */
func KickPlayer(srv *Server, cl int) {
    cmd := fmt.Sprintf("kick %d", cl)
    WriteByte(SCMDCommand, &srv.messageout)
    WriteString(cmd, &srv.messageout)

    txt := fmt.Sprintf("KICK [%d] was kicked", cl)
    LogEventToDatabase(srv.id, LogTypeCommand, txt)
}
