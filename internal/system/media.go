package system

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"nex-server/internal/models"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/godbus/dbus/v5"
)

type MediaController struct {
	conn *dbus.Conn
	uid  int
}

func NewMediaController() *MediaController {
	uid := getRealUserID()
	
	busAddress := fmt.Sprintf("unix:path=/run/user/%d/bus", uid)
	
	conn, err := dbus.Connect(busAddress)
	if err != nil {
		return &MediaController{uid: uid}
	}
	return &MediaController{conn: conn, uid: uid}
}

func getRealUserID() int {
	sudoUID := os.Getenv("SUDO_UID")
	if sudoUID != "" {
		if uid, err := strconv.Atoi(sudoUID); err == nil {
			return uid
		}
	}
	
	pkexecUID := os.Getenv("PKEXEC_UID")
	if pkexecUID != "" {
		if uid, err := strconv.Atoi(pkexecUID); err == nil {
			return uid
		}
	}
	
	uid := os.Getuid()
	if uid == 0 {
		files, err := ioutil.ReadDir("/run/user")
		if err == nil {
			for _, f := range files {
				if f.IsDir() && f.Name() != "0" {
					if id, err := strconv.Atoi(f.Name()); err == nil && id >= 1000 {
						return id
					}
				}
			}
		}
		return 1000
	}
	return uid
}

func (m *MediaController) GetAllStatus() []models.AudioState {
	states := m.getStatusViaPlayerctl()
	if len(states) > 0 {
		return states
	}
	
	return m.getStatusViaDbus()
}

func (m *MediaController) getStatusViaPlayerctl() []models.AudioState {
	username := m.getUsername()
	
	cmd := exec.Command("runuser", "-u", username, "--", "env", fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%d", m.uid), "playerctl", "-l")
	out, err := cmd.Output()
	if err != nil {
		return []models.AudioState{}
	}
	
	players := strings.Split(strings.TrimSpace(string(out)), "\n")
	var states []models.AudioState
	
	playerCount := 0
	for _, player := range players {
		if player == "" {
			continue
		}
		
		state := m.getPlayerInfo(player, username)
		if state.Title != "" {
			playerCount++
			state.ID = fmt.Sprintf("player%d", playerCount)
			
			if strings.Contains(player, "youtube_music") {
				state.Name = "Youtube Music"
			} else if strings.Contains(strings.ToLower(player), "firefox") {
				state.Name = "Firefox"
			} else if strings.Contains(strings.ToLower(player), "spotify") {
				state.Name = "Spotify"
			} else if strings.Contains(strings.ToLower(player), "chrome") || strings.Contains(strings.ToLower(player), "chromium") {
				state.Name = "Chrome"
			} else if strings.Contains(strings.ToLower(player), "vlc") {
				state.Name = "VLC"
			} else {
				state.Name = player
			}

			if strings.HasPrefix(state.ArtURL, "file://") {
				path := strings.TrimPrefix(state.ArtURL, "file://")
				encoded := base64.URLEncoding.EncodeToString([]byte(path))
				state.ArtURL = fmt.Sprintf("/v1/img/tmp/%s", encoded)
			}

			states = append(states, state)
		}
	}
	
	return states
}

func (m *MediaController) getUsername() string {
	data, err := ioutil.ReadFile("/etc/passwd")
	if err != nil {
		return fmt.Sprintf("#%d", m.uid)
	}
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			if parts[2] == strconv.Itoa(m.uid) {
				return parts[0]
			}
		}
	}
	
	return fmt.Sprintf("#%d", m.uid)
}

func (m *MediaController) getPlayerInfo(player string, username string) models.AudioState {
	runCmd := func(args ...string) string {
		cmdArgs := append([]string{"-u", username, "--", "env", fmt.Sprintf("XDG_RUNTIME_DIR=/run/user/%d", m.uid), "playerctl", "-p", player}, args...)
		cmd := exec.Command("runuser", cmdArgs...)
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	
	status := runCmd("status")
	playing := status == "Playing"
	
	title := runCmd("metadata", "title")
	artist := runCmd("metadata", "artist")
	album := runCmd("metadata", "album")
	artUrl := runCmd("metadata", "mpris:artUrl")
	
	var position int64 = 0
	posStr := runCmd("position")
	if posStr != "" {
		val, _ := strconv.ParseFloat(posStr, 64)
		position = int64(val)
	}
	
	var duration int64 = 0
	durStr := runCmd("metadata", "mpris:length")
	if durStr != "" {
		val, _ := strconv.ParseInt(durStr, 10, 64)
		duration = val / 1000000
	}
	
	return models.AudioState{
		Playing:   playing,
		Artist:    artist,
		Title:     title,
		Album:     album,
		ArtURL:    artUrl,
		Timestamp: position,
		Duration:  duration,
	}
}

func (m *MediaController) getStatusViaDbus() []models.AudioState {
	if m.conn == nil {
		return []models.AudioState{}
	}

	names, err := m.listNames()
	if err != nil {
		return []models.AudioState{}
	}

	var states []models.AudioState

	for _, name := range names {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2.") {
			obj := m.conn.Object(name, "/org/mpris/MediaPlayer2")
			
			statusVar, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
			if err != nil {
				continue
			}
			
			statusStr, ok := statusVar.Value().(string)
			if !ok {
				continue
			}
			playing := statusStr == "Playing"
			
			metadataVar, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
			if err != nil {
				continue
			}

			metadata, ok := metadataVar.Value().(map[string]dbus.Variant)
			if !ok {
				continue
			}
			
			artist := getStringFromMetadata(metadata, "xesam:artist")
			title := getStringFromMetadata(metadata, "xesam:title")
			album := getStringFromMetadata(metadata, "xesam:album")
			artUrl := getStringFromMetadata(metadata, "mpris:artUrl")
			duration := getInt64FromMetadata(metadata, "mpris:length")
			
			posVar, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Position")
			var position int64 = 0
			if err == nil {
				if pos, ok := posVar.Value().(int64); ok {
					position = pos
				}
			}

			if title != "" {
				states = append(states, models.AudioState{
					Playing:   playing,
					Artist:    artist,
					Title:     title,
					Album:     album,
					ArtURL:    artUrl,
					Timestamp: position / 1000000, 
					Duration:  duration / 1000000,
				})
			}
		}
	}

	return states
}

func (m *MediaController) GetStatus() models.AudioState {
	if m.conn == nil {
		return models.AudioState{}
	}

	names, err := m.listNames()
	if err != nil {
		return models.AudioState{}
	}

	var bestState models.AudioState
	var bestScore int = -1

	for _, name := range names {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2.") {
			obj := m.conn.Object(name, "/org/mpris/MediaPlayer2")
			
			statusVar, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
			if err != nil {
				continue
			}
			
			statusStr := statusVar.Value().(string)
			playing := statusStr == "Playing"
			
			score := 0
			if strings.Contains(name, "youtube_music") {
				score += 100
			}
			if strings.Contains(name, "spotify") {
				score += 90
			}
			if playing {
				score += 50
			}

			if score > bestScore {
				metadataVar, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
				if err != nil {
					continue
				}

				metadata := metadataVar.Value().(map[string]dbus.Variant)
				
				artist := getStringFromMetadata(metadata, "xesam:artist")
				title := getStringFromMetadata(metadata, "xesam:title")
				album := getStringFromMetadata(metadata, "xesam:album")
				artUrl := getStringFromMetadata(metadata, "mpris:artUrl")
				duration := getInt64FromMetadata(metadata, "mpris:length")
				
				posVar, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Position")
				var position int64 = 0
				if err == nil {
					position = int64(posVar.Value().(int64))
				}

				bestState = models.AudioState{
					Playing:   playing,
					Artist:    artist,
					Title:     title,
					Album:     album,
					ArtURL:    artUrl,
					Timestamp: position / 1000000, 
					Duration:  duration / 1000000,
				}
				bestScore = score
			}
		}
	}

	if bestScore != -1 {
		return bestState
	}

	return models.AudioState{}
}

func getStringFromMetadata(m map[string]dbus.Variant, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.Value().(string); ok {
			return str
		}
		if strs, ok := val.Value().([]string); ok && len(strs) > 0 {
			return strs[0]
		}
	}
	return ""
}

func getInt64FromMetadata(m map[string]dbus.Variant, key string) int64 {
	if val, ok := m[key]; ok {
		if i, ok := val.Value().(int64); ok {
			return i
		}
		if i, ok := val.Value().(uint64); ok {
			return int64(i)
		}
	}
	return 0
}

func (m *MediaController) PlayPause() {
	m.callMethod("PlayPause")
}

func (m *MediaController) Next() {
	m.callMethod("Next")
}

func (m *MediaController) Previous() {
	m.callMethod("Previous")
}

func (m *MediaController) SetPosition(position int64) {
	if m.conn == nil { return }
	names, _ := m.listNames()
	
	var bestPlayer string
	var bestScore int = -1

	for _, name := range names {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2.") {
			obj := m.conn.Object(name, "/org/mpris/MediaPlayer2")
			statusVar, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
			playing := false
			if err == nil {
				statusStr := statusVar.Value().(string)
				playing = statusStr == "Playing"
			}

			score := 0
			if strings.Contains(name, "youtube_music") {
				score += 100
			}
			if playing {
				score += 50
			}
			
			if score > bestScore {
				bestScore = score
				bestPlayer = name
			}
		}
	}

	if bestPlayer != "" {
		obj := m.conn.Object(bestPlayer, "/org/mpris/MediaPlayer2")
		metadataVar, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
		if err != nil {
			return
		}
		metadata := metadataVar.Value().(map[string]dbus.Variant)
		
		var trackID dbus.ObjectPath
		if val, ok := metadata["mpris:trackid"]; ok {
			if path, ok := val.Value().(dbus.ObjectPath); ok {
				trackID = path
			} else if str, ok := val.Value().(string); ok {
				trackID = dbus.ObjectPath(str)
			}
		}

		if trackID != "" {
			targetMicros := position * 1000
			obj.Call("org.mpris.MediaPlayer2.Player.SetPosition", 0, trackID, targetMicros)
		}
	}
}

func (m *MediaController) callMethod(method string) {
	if m.conn == nil { return }
	names, _ := m.listNames()
	
	var bestPlayer string
	var bestScore int = -1

	for _, name := range names {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2.") {
			obj := m.conn.Object(name, "/org/mpris/MediaPlayer2")
			statusVar, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
			playing := false
			if err == nil {
				statusStr := statusVar.Value().(string)
				playing = statusStr == "Playing"
			}

			score := 0
			if strings.Contains(name, "youtube_music") {
				score += 100
			}
			if playing {
				score += 50
			}
			
			if score > bestScore {
				bestScore = score
				bestPlayer = name
			}
		}
	}

	if bestPlayer != "" {
		obj := m.conn.Object(bestPlayer, "/org/mpris/MediaPlayer2")
		obj.Call("org.mpris.MediaPlayer2.Player."+method, 0)
	}
}

func (m *MediaController) listNames() ([]string, error) {
	var names []string
	err := m.conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	return names, err
}
