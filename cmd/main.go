package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/j6s/harbour-whisperfish-message-restore/lib"
	"github.com/j6s/harbour-whisperfish-message-restore/lib/transfer"
)

var SignalDesktopFile = flag.String("signal-desktop-file", "", "Path to SQLite DB from signal desktop")
var WhisperfishFile = flag.String("whisperfish-file", "", "Path to SQLite DB from whisperfish")

func main() {
	flag.Parse()
	if *SignalDesktopFile == "" {
		log.Printf("--signal-desktop-file is required")
		flag.PrintDefaults()
		return
	}
	if *WhisperfishFile == "" {
		log.Printf("--whisperfish-file is required")
		flag.PrintDefaults()
		return
	}

	t, err := transfer.CreateTransfer(*SignalDesktopFile, *WhisperfishFile)
	if err != nil {
		log.Fatal(err)
	}

	privateMessages, err := t.FetchPrivateMessageTransferSets()
	if err != nil {
		log.Fatal(err)
	}
	//groupMessages, err := t.FetchGroupMessageTransferSets()
	//if err != nil {
	//	log.Fatal(err)
	//}

	if askTransferPrivateMessages(privateMessages) {
		results := t.TransferPrivateMessages(privateMessages)
		for _, result := range results {
			log.Printf("Private Messages to %s (%s): %v", result.Set.Recipient.FullName, result.Set.Recipient.ServiceId, result.Success)
			if result.Message != "" {
				log.Printf("\t%s", result.Message)
			}
		}
	} else {
		log.Printf("Not importing messages")
	}

	//if askTransferGroupMessages(groupMessages) {
	//	results := t.TransferGroupMessages(groupMessages)
	//	for _, result := range results {
	//		log.Printf("Group Message %s: %v", result.Set.Group.Name, result.Success)
	//		if result.Message != "" {
	//			log.Printf("\t%s", result.Message)
	//		}
	//	}
	//}
}

func askTransferPrivateMessages(sets []transfer.PrivateMessageTransferSet) bool {
	contacts := len(sets)
	conversations := 0
	messages := 0
	for _, set := range sets {
		if len(set.Messages) > 0 {
			conversations += 1
			messages += len(set.Messages)
		}
	}

	return lib.AskConfirm(fmt.Sprintf("Continue transferring %d contacts, %d conversations, %d messages?", contacts, conversations, messages))
}

func askTransferGroupMessages(sets []transfer.GroupMessageTransferSet) bool {
	groups := len(sets)
	messages := 0
	for _, set := range sets {
		messages += len(set.Messages)
	}

	return lib.AskConfirm(fmt.Sprintf("Continue transferring %d groups, %d messages?", groups, messages))
}
