package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

var (
	Token              string
	HWND               syscall.Handle
	user32             = syscall.MustLoadDLL("user32.dll")
	procEnumWindows    = user32.MustFindProc("EnumWindows")
	procGetWindowTextW = user32.MustFindProc("GetWindowTextW")
	sendMessage        = user32.MustFindProc("SendMessageW")
)
var ()

func init() {
	flag.StringVar(&Token, "token", "", "Bot token")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)

	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents |= discordgo.IntentsMessageContent

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	err = dg.Close()
	if err != nil {
		fmt.Println("error closing dg,", err)
		return
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate, hwnd syscall.Handle) {
	fmt.Println("message received")
	if m.Author.ID == s.State.User.ID {
		return
	}
	HWND, err := FindWindow("Discord")
	if err != nil {
		fmt.Println("error finding discord window,", err)
		return
	}
	fmt.Println("hwnd:", HWND)
	a, b, err = sendMessage.Call(uintptr(HWND), 0x100, 0x41, 0x1)
	if err != nil {
		fmt.Println("error sending message,", err)
		return
	}
	if m.Content == "ping" {
		fmt.Println("attempting to send message")
		s.ChannelMessageSend(m.ChannelID, "pong")
	}

	if m.Content == "pong" {
		fmt.Println("attempting to send message")
		s.ChannelMessageSend(m.ChannelID, "ping")
	}
}

// thanks github! https://gist.github.com/EliCDavis/5374fa4947897b16a81f6550d142ab28
func EnumWindows(enumFunc uintptr, lparam uintptr) (err error) {
	r1, _, e1 := syscall.SyscallN(procEnumWindows.Addr(), 2, uintptr(enumFunc), uintptr(lparam), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func GetWindowText(hwnd syscall.Handle, str *uint16, maxCount int32) (len int32, err error) {
	r0, _, e1 := syscall.SyscallN(procGetWindowTextW.Addr(), 3, uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	len = int32(r0)
	if len == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func FindWindow(title string) (syscall.Handle, error) {
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			// ignore the error
			return 1 // continue enumeration
		}
		if syscall.UTF16ToString(b) == title {
			// note the window
			HWND = h
			return 0 // stop enumeration
		}
		return 1 // continue enumeration
	})
	err = EnumWindows(cb, 0)
	if err != nil {
		fmt.Println("error enumerating windows:", err)
	}
	if HWND == 0 {
		return 0, fmt.Errorf("No window with title '%s' found", title)
	}
	return HWND, nil
}
