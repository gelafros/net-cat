package client

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/jroimartin/gocui"
)

const (
	viewMessages     = "messages"
	viewInput        = "input"
	viewClientName   = "client-name"
	viewStatus       = "status"
	viewUsers        = "users"
	viewRooms        = "rooms"
	viewMouseCapture = "mouse-capture"
)

type app struct {
	gui          *gocui.Gui
	conn         net.Conn
	addr         string
	mu           sync.Mutex
	closed       bool
	follow       bool
	name         string
	inputHistory []string
	historyIndex int
	users        map[string]bool
	currentRoom  string
	rooms        []string
}

var (
	joinPattern         = regexp.MustCompile(`User \[([^\]]+)\] has joined our chat`)
	leavePattern        = regexp.MustCompile(`User \[([^\]]+)\] has left our chat`)
	renamePattern       = regexp.MustCompile(`\[([^\]]+)\] changed their name to \[([^\]]+)\]`)
	welcomeNamePattern  = regexp.MustCompile(`>>>\s*([^<]+?)\s*<<<`)
	serverPromptPattern = regexp.MustCompile(`(?:\x1b\[[0-9;]*m)*\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]\[[^\]\n]+\]:(?:\x1b\[[0-9;]*m)*(?:\n|$)`)
	joinedRoomPattern   = regexp.MustCompile(`---\s*Joined Room:\s*(.+?)\s*---`)
	returnedMainPattern = regexp.MustCompile(`---\s*Returned to Main\s*---`)
	availableRoomLine   = regexp.MustCompile(`(?m)^-\s*(.+?\(\d+\s*users\))$`)
	roomsBlockPattern   = regexp.MustCompile(`(?m)^Available rooms:\s*\n(?:-\s*.+?\(\d+\s*users\)\s*\n?)*`)
)

func Run(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to %s: %v\n", addr, err)
		os.Exit(1)
	}
	defer conn.Close()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start ui: %v\n", err)
		os.Exit(1)
	}
	defer g.Close()

	a := &app{
		gui:          g,
		conn:         conn,
		addr:         addr,
		follow:       true,
		historyIndex: -1,
		users:        make(map[string]bool),
		currentRoom:  "Main",
	}

	g.Cursor = true
	g.Mouse = true
	g.SetManagerFunc(a.layout)

	if err := a.bindKeys(); err != nil {
		fmt.Fprintf(os.Stderr, "bind keys: %v\n", err)
		os.Exit(1)
	}

	go a.readServer()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		fmt.Fprintf(os.Stderr, "ui loop: %v\n", err)
		os.Exit(1)
	}
}

func (a *app) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if maxX < 24 || maxY < 10 {
		return nil
	}

	if v, err := g.SetView(viewMouseCapture, -1, -1, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		if _, err := g.SetViewOnBottom(viewMouseCapture); err != nil {
			return err
		}
	}

	roomsWidth := 0
	if maxX >= 80 {
		roomsWidth = 20
	}
	usersWidth := 0
	if maxX >= 60 {
		usersWidth = 22
	}
	messagesLeft := 0
	if roomsWidth > 0 {
		messagesLeft = roomsWidth + 1
	}
	messagesRight := maxX - usersWidth - 1

	if roomsWidth > 0 {
		if v, err := g.SetView(viewRooms, 0, 0, roomsWidth, maxY-7); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " Rooms "
			v.Wrap = true
		}
		a.renderRooms()
	} else {
		g.DeleteView(viewRooms)
	}

	if v, err := g.SetView(viewMessages, messagesLeft, 0, messagesRight, maxY-7); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Net-Cat "
		if a.currentRoom != "" {
			v.Title = fmt.Sprintf(" Room: %s ", a.currentRoom)
		}
		v.Wrap = true
		v.Autoscroll = true
		fmt.Fprintln(v, "Connected to "+a.addr)
		fmt.Fprintln(v, "Type your chat name below and press Enter.")
	}

	if usersWidth > 0 {
		if v, err := g.SetView(viewUsers, messagesRight+1, 0, maxX-1, maxY-7); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = " Users "
			v.Wrap = true
		}
		a.renderUsers()
	} else {
		g.DeleteView(viewUsers)
	}

	if v, err := g.SetView(viewClientName, 0, maxY-6, maxX-1, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}
	a.renderClientName()

	if v, err := g.SetView(viewInput, 0, maxY-4, maxX-1, maxY-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Type Message "
		v.Editable = true
		v.Wrap = true
		if _, err := g.SetCurrentView(viewInput); err != nil {
			return err
		}
	}

	if v, err := g.SetView(viewStatus, 0, maxY-2, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}

	a.drawStatus("Enter sends | PgUp/PgDn scroll | End follows latest | Ctrl+C quits")
	return nil
}

func (a *app) bindKeys() error {
	if err := a.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, a.quit); err != nil {
		return err
	}
	if err := a.gui.SetKeybinding(viewInput, gocui.KeyEnter, gocui.ModNone, a.sendMessage); err != nil {
		return err
	}
	if err := a.gui.SetKeybinding(viewInput, gocui.KeyArrowUp, gocui.ModNone, a.previousInput); err != nil {
		return err
	}
	if err := a.gui.SetKeybinding(viewInput, gocui.KeyArrowDown, gocui.ModNone, a.nextInput); err != nil {
		return err
	}
	if err := a.gui.SetKeybinding("", gocui.KeyPgup, gocui.ModNone, a.scrollMessagesUp); err != nil {
		return err
	}
	if err := a.gui.SetKeybinding("", gocui.KeyPgdn, gocui.ModNone, a.scrollMessagesDown); err != nil {
		return err
	}
	if err := a.gui.SetKeybinding("", gocui.KeyHome, gocui.ModNone, a.scrollMessagesTop); err != nil {
		return err
	}
	if err := a.gui.SetKeybinding("", gocui.KeyEnd, gocui.ModNone, a.scrollMessagesBottom); err != nil {
		return err
	}
	for _, viewName := range []string{viewMessages, viewInput, viewClientName, viewUsers, viewRooms, viewMouseCapture} {
		if err := a.gui.SetKeybinding(viewName, gocui.MouseWheelUp, gocui.ModNone, a.scrollMessagesWheelUp); err != nil {
			return err
		}
		if err := a.gui.SetKeybinding(viewName, gocui.MouseWheelDown, gocui.ModNone, a.scrollMessagesWheelDown); err != nil {
			return err
		}
	}
	return nil
}

func (a *app) sendMessage(g *gocui.Gui, v *gocui.View) error {
	text := strings.TrimSpace(v.Buffer())
	v.Clear()
	if err := v.SetCursor(0, 0); err != nil {
		return err
	}
	if err := v.SetOrigin(0, 0); err != nil {
		return err
	}
	if text == "" {
		a.drawStatus("empty message skipped")
		return nil
	}

	if _, err := fmt.Fprintln(a.conn, text); err != nil {
		a.drawStatus("send failed: " + err.Error())
		return nil
	}

	a.rememberInput(text)
	a.drawStatus("sent at " + time.Now().Format("15:04:05"))
	return nil
}

func (a *app) previousInput(g *gocui.Gui, v *gocui.View) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.inputHistory) == 0 {
		return nil
	}
	if a.historyIndex == -1 {
		a.historyIndex = len(a.inputHistory) - 1
	} else if a.historyIndex > 0 {
		a.historyIndex--
	}
	return replaceInput(v, a.inputHistory[a.historyIndex])
}

func (a *app) nextInput(g *gocui.Gui, v *gocui.View) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.inputHistory) == 0 || a.historyIndex == -1 {
		return nil
	}
	a.historyIndex++
	if a.historyIndex >= len(a.inputHistory) {
		a.historyIndex = -1
		return replaceInput(v, "")
	}
	return replaceInput(v, a.inputHistory[a.historyIndex])
}

func (a *app) rememberInput(text string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if text == "/rooms" || strings.HasPrefix(text, "/join") || strings.HasPrefix(text, "/create") || text == "/leave" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			fmt.Fprintln(a.conn, "/rooms")
		}()
	}
	if len(a.inputHistory) == 0 || a.inputHistory[len(a.inputHistory)-1] != text {
		a.inputHistory = append(a.inputHistory, text)
	}
	a.historyIndex = -1
}

func (a *app) readServer() {
	buf := make([]byte, 4096)
	for {
		n, err := a.conn.Read(buf)
		if n > 0 {
			text := cleanTerminalText(string(buf[:n]))
			if text != "" {
				a.gui.Update(func(g *gocui.Gui) error {
					a.updateUsers(text)
					a.updateRoomsAndTitle(text)
					displayText := removeServerPrompts(text)
					if displayText == "" {
						return nil
					}
					v, err := g.View(viewMessages)
					if err != nil {
						return nil
					}
					fmt.Fprint(v, displayText)
					if a.shouldFollow() {
						scrollViewToBottom(v)
					}
					return nil
				})
			}
		}
		if err != nil {
			a.gui.Update(func(g *gocui.Gui) error {
				a.drawStatus("connection closed: " + err.Error())
				return gocui.ErrQuit
			})
			return
		}
	}
}

func (a *app) scrollMessagesUp(g *gocui.Gui, _ *gocui.View) error {
	v, err := g.View(viewMessages)
	if err != nil {
		return nil
	}
	_, oy := v.Origin()
	_, height := v.Size()
	step := max(1, height-1)
	v.Autoscroll = false
	if err := v.SetOrigin(0, max(0, oy-step)); err != nil {
		return nil
	}
	a.setFollow(false)
	a.drawStatus("history: PageDown moves down, End jumps to latest")
	return nil
}

func (a *app) scrollMessagesDown(g *gocui.Gui, _ *gocui.View) error {
	v, err := g.View(viewMessages)
	if err != nil {
		return nil
	}
	_, oy := v.Origin()
	_, height := v.Size()
	step := max(1, height-1)
	bottom := bottomOrigin(v)
	next := min(bottom, oy+step)
	if err := v.SetOrigin(0, next); err != nil {
		return nil
	}
	if next >= bottom {
		v.Autoscroll = true
		a.setFollow(true)
		a.drawStatus("following latest messages")
		return nil
	}
	v.Autoscroll = false
	a.setFollow(false)
	a.drawStatus("history: PageDown moves down, End jumps to latest")
	return nil
}

func (a *app) scrollMessagesTop(g *gocui.Gui, _ *gocui.View) error {
	v, err := g.View(viewMessages)
	if err != nil {
		return nil
	}
	v.Autoscroll = false
	if err := v.SetOrigin(0, 0); err != nil {
		return nil
	}
	a.setFollow(false)
	a.drawStatus("top of history")
	return nil
}

func (a *app) scrollMessagesBottom(g *gocui.Gui, _ *gocui.View) error {
	v, err := g.View(viewMessages)
	if err != nil {
		return nil
	}
	scrollViewToBottom(v)
	a.setFollow(true)
	a.drawStatus("following latest messages")
	return nil
}

func (a *app) scrollMessagesWheelUp(g *gocui.Gui, _ *gocui.View) error {
	return a.scrollMessagesBy(g, -3)
}

func (a *app) scrollMessagesWheelDown(g *gocui.Gui, _ *gocui.View) error {
	return a.scrollMessagesBy(g, 3)
}

func (a *app) scrollMessagesBy(g *gocui.Gui, delta int) error {
	v, err := g.View(viewMessages)
	if err != nil {
		return nil
	}
	_, oy := v.Origin()
	bottom := bottomOrigin(v)
	next := min(bottom, max(0, oy+delta))
	if err := v.SetOrigin(0, next); err != nil {
		return nil
	}
	if next >= bottom {
		v.Autoscroll = true
		a.setFollow(true)
		a.drawStatus("following latest messages")
		return nil
	}
	v.Autoscroll = false
	a.setFollow(false)
	a.drawStatus("history: mouse wheel/PageDown moves down, End jumps to latest")
	return nil
}

func (a *app) shouldFollow() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.follow
}

func (a *app) setFollow(follow bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.follow = follow
}

func (a *app) updateUsers(text string) {
	plain := stripANSI(text)

	a.mu.Lock()
	defer a.mu.Unlock()

	for _, match := range joinPattern.FindAllStringSubmatch(plain, -1) {
		a.users[strings.TrimSpace(match[1])] = true
	}
	for _, match := range welcomeNamePattern.FindAllStringSubmatch(plain, -1) {
		name := strings.TrimSpace(match[1])
		if name != "" {
			a.name = name
			a.users[name] = true
			go func() {
				time.Sleep(500 * time.Millisecond)
				fmt.Fprintln(a.conn, "/rooms")
			}()
		}
	}
	for _, match := range renamePattern.FindAllStringSubmatch(plain, -1) {
		oldName := strings.TrimSpace(match[1])
		newName := strings.TrimSpace(match[2])
		delete(a.users, oldName)
		if newName != "" {
			if oldName == a.name {
				a.name = newName
			}
			a.users[newName] = true
		}
	}
	for _, match := range leavePattern.FindAllStringSubmatch(plain, -1) {
		delete(a.users, strings.TrimSpace(match[1]))
	}

	a.renderUsersLocked()
	a.renderClientNameLocked()
}

func (a *app) updateRoomsAndTitle(text string) {
	plain := stripANSI(text)

	a.mu.Lock()
	defer a.mu.Unlock()

	joinedMatches := joinedRoomPattern.FindAllStringSubmatch(plain, -1)
	if len(joinedMatches) > 0 {
		a.currentRoom = strings.TrimSpace(joinedMatches[len(joinedMatches)-1][1])
	} else if returnedMainPattern.MatchString(plain) {
		a.currentRoom = "Main"
	}

	roomLines := availableRoomLine.FindAllStringSubmatch(plain, -1)
	if len(roomLines) > 0 {
		a.rooms = nil
		for _, m := range roomLines {
			a.rooms = append(a.rooms, strings.TrimSpace(m[1]))
		}
	}

	a.renderRoomsLocked()
	a.updateChatTitleLocked()
}

func (a *app) updateChatTitleLocked() {
	v, err := a.gui.View(viewMessages)
	if err != nil {
		return
	}
	if a.currentRoom == "" {
		v.Title = " Net-Cat "
	} else {
		v.Title = fmt.Sprintf(" Room: %s ", a.currentRoom)
	}
}

func (a *app) clientNameLabelLocked() string {
	if a.name == "" {
		return ""
	}
	return a.name
}

func (a *app) renderClientName() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.renderClientNameLocked()
}

func (a *app) renderClientNameLocked() {
	v, err := a.gui.View(viewClientName)
	if err != nil {
		return
	}
	v.Clear()
	label := a.clientNameLabelLocked()
	if label == "" {
		return
	}
	width, _ := v.Size()
	fmt.Fprint(v, "\033[1m"+centerText(label, width)+"\033[0m")
}

func (a *app) renderUsers() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.renderUsersLocked()
}

func (a *app) renderUsersLocked() {
	v, err := a.gui.View(viewUsers)
	if err != nil {
		return
	}
	v.Clear()
	if len(a.users) == 0 {
		fmt.Fprintln(v, "Waiting...")
		return
	}

	names := make([]string, 0, len(a.users))
	for name := range a.users {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Fprintln(v, name)
	}
}

func (a *app) renderRooms() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.renderRoomsLocked()
}

func (a *app) renderRoomsLocked() {
	v, err := a.gui.View(viewRooms)
	if err != nil {
		return
	}
	v.Clear()
	if len(a.rooms) == 0 {
		fmt.Fprintln(v, "Waiting...")
		return
	}
	for _, r := range a.rooms {
		fmt.Fprintln(v, r)
	}
}

func (a *app) drawStatus(text string) {
	v, err := a.gui.View(viewStatus)
	if err != nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	v.Clear()
	if a.closed {
		fmt.Fprint(v, "closed")
		return
	}
	fmt.Fprint(v, text)
}

func (a *app) quit(g *gocui.Gui, v *gocui.View) error {
	a.mu.Lock()
	a.closed = true
	a.mu.Unlock()
	_ = a.conn.Close()
	return gocui.ErrQuit
}

func cleanTerminalText(s string) string {
	var out strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\x1b' {
			if i+1 < len(runes) && runes[i+1] == '[' {
				seq := []rune{r, runes[i+1]}
				i += 2
				for i < len(runes) {
					seq = append(seq, runes[i])
					if isANSIFinal(runes[i]) {
						break
					}
					i++
				}
				if len(seq) > 0 && seq[len(seq)-1] == 'm' {
					out.WriteString(string(seq))
				}
			} else if i+1 < len(runes) {
				i++
			}
			continue
		}
		if r == '\r' {
			continue
		}
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func removeServerPrompts(s string) string {
	s = serverPromptPattern.ReplaceAllString(s, "")
	s = roomsBlockPattern.ReplaceAllString(s, "")
	return s
}

func replaceInput(v *gocui.View, text string) error {
	v.Clear()
	fmt.Fprint(v, text)
	if err := v.SetOrigin(0, 0); err != nil {
		return err
	}
	x := len([]rune(text))
	maxX, _ := v.Size()
	if maxX > 0 && x >= maxX {
		x = maxX - 1
	}
	return v.SetCursor(x, 0)
}

func centerText(text string, width int) string {
	if width <= 0 {
		return text
	}
	textWidth := len([]rune(text))
	if textWidth >= width {
		return text
	}
	return strings.Repeat(" ", (width-textWidth)/2) + text
}

func scrollViewToBottom(v *gocui.View) {
	v.Autoscroll = true
	_ = v.SetOrigin(0, bottomOrigin(v))
}

func bottomOrigin(v *gocui.View) int {
	_, height := v.Size()
	return max(0, len(v.BufferLines())-height)
}

func stripANSI(s string) string {
	var out strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\x1b' {
			if i+1 < len(runes) && runes[i+1] == '[' {
				i += 2
				for i < len(runes) && !isANSIFinal(runes[i]) {
					i++
				}
			} else if i+1 < len(runes) {
				i++
			}
			continue
		}
		if r == '\r' {
			continue
		}
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func isANSIFinal(r rune) bool {
	return r >= 0x40 && r <= 0x7e
}
