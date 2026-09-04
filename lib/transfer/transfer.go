package transfer

import (
	"fmt"
	"log"

	"github.com/j6s/harbour-whisperfish-message-restore/lib/data"
	"github.com/j6s/harbour-whisperfish-message-restore/lib/signalDesktop"
	"github.com/j6s/harbour-whisperfish-message-restore/lib/whisperfish"
)

type PrivateMessageTransferSet struct {
	Me        data.User
	Recipient data.User
	Messages  []data.Message
}

type PrivateMessageTransferResult struct {
	Set     PrivateMessageTransferSet
	Success bool
	Message string
}

type GroupMessageTransferSet struct {
	Me       data.User
	Group    data.Group
	Messages []data.Message
}

type GroupMessageTransferResult struct {
	Set     GroupMessageTransferSet
	Success bool
	Message string
}

type Transfer struct {
	Signal  signalDesktop.SignalDesktop
	Whisper whisperfish.Whisperfish
}

func CreateTransfer(signalDatabaseLocation string, whisperDatabaseLocation string) (*Transfer, error) {
	signal, err := signalDesktop.CreateSignalDesktop(signalDatabaseLocation)
	if err != nil {
		return nil, fmt.Errorf("CreateTransfer: Cannot create signal instance: %v", err)
	}

	whisper, err := whisperfish.CreateWhisperfish(whisperDatabaseLocation)
	if err != nil {
		return nil, fmt.Errorf("CreateTransfer: Cannot create whisper instance: %v", err)
	}

	return &Transfer{
		Signal:  *signal,
		Whisper: *whisper,
	}, nil
}

func (self *Transfer) FetchPrivateMessageTransferSets() ([]PrivateMessageTransferSet, error) {
	var sets []PrivateMessageTransferSet
	users, err := self.Signal.GetUsers()
	if err != nil {
		return sets, fmt.Errorf("Transfer.FetchPrivateMessageTransferSets: Failed to get users: %v", err)
	}

	me, err := self.Signal.FindMe(users)
	if err != nil {
		return sets, fmt.Errorf("Transfer.FetchPrivateMessageTransferSets: Failed to find me: %v", err)
	}

	for _, user := range users.All {
		if user.ServiceId == "" {
			continue
		}
		messages, err := self.Signal.GetPrivateMessages(user, me)
		if err != nil {
			return sets, fmt.Errorf("Transfer.FetchPrivateMessageTransferSets: Failed to get messages: %v", err)
		}

		sets = append(sets, PrivateMessageTransferSet{
			Me:        me,
			Recipient: user,
			Messages:  messages,
		})
	}

	return sets, nil
}

func (self *Transfer) TransferPrivateMessages(sets []PrivateMessageTransferSet) []PrivateMessageTransferResult {
	results := make([]PrivateMessageTransferResult, len(sets))
	for i, set := range sets {
		results[i] = self.TransferPrivateMessage(set)
	}

	return results
}

func (self *Transfer) TransferPrivateMessage(set PrivateMessageTransferSet) PrivateMessageTransferResult {
	result := PrivateMessageTransferResult{Success: true, Set: set}

	recipient, err := self.Whisper.EnsureRecipientExist(set.Recipient)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Transfer.TranferPrivateMessage: Failed to ensure recipient: %v", err)
		return result
	}

	me, err := self.Whisper.EnsureRecipientExist(set.Me)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Transfer.TranferPrivateMessage: Failed to ensure me: %v", err)
		return result
	}

	if len(set.Messages) == 0 {
		result.Success = true
		return result
	}

	session, err := self.Whisper.EnsurePrivateMessageSessionExists(recipient)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Transfer.TranferPrivateMessage: Failed to ensure private message session: %v", err)
		return result
	}

	for _, message := range set.Messages {
		_, err = self.Whisper.EnsureMessageExists(me, recipient, session, message)
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Transfer.TranferPrivateMessage: %s, Failed to ensure message %#v exists: %v", result.Message, message, err)
		}
	}

	return result
}

func (self *Transfer) FetchGroupMessageTransferSets() ([]GroupMessageTransferSet, error) {
	var sets []GroupMessageTransferSet
	users, err := self.Signal.GetUsers()
	if err != nil {
		return sets, fmt.Errorf("Transfer.FetchGroupMessageTransferSets: Failed to get users: %v", err)
	}

	me, err := self.Signal.FindMe(users)
	if err != nil {
		return sets, fmt.Errorf("Transfer.FetchGroupMessageTransferSets: ailed to find me: %v", err)
	}

	groups, err := self.Signal.GetGroups(users)
	if err != nil {
		return sets, fmt.Errorf("Transfer.FetchGroupMessageTransferSets: Failed to get groups: %v", err)
	}

	for _, group := range groups {
		messages, err := self.Signal.GetGroupMessages(group, me)
		if err != nil {
			return sets, fmt.Errorf("Transfer.FetchGroupMessageTransferSets: Failed to get messages: %v", err)
		}

		sets = append(sets, GroupMessageTransferSet{
			Me:       me,
			Group:    group,
			Messages: messages,
		})
	}

	return sets, nil
}

func (self *Transfer) TransferGroupMessage(set GroupMessageTransferSet) GroupMessageTransferResult {
	result := GroupMessageTransferResult{Success: true, Set: set}

	whisperGroup, err := self.Whisper.EnsureGroupExists(set.Group)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Transfer.TransferGroupMessage: Failed to ensure group exists: %v", err)
		return result
	}

	err = self.Whisper.EnsureGroupMembersExist(set.Group, whisperGroup)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Transfer.TransferGroupMessage: Failed to ensure group members exists: %v", err)
		return result
	}

	session, err := self.Whisper.EnsureGroupMessageSessionExists(whisperGroup)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Transfer.TransferGroupMessage: Failed to ensure group message session: %v", err)
		return result
	}

	me, err := self.Whisper.EnsureRecipientExist(set.Me)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Transfer.TransferGroupMessage: Failed to ensure 'me' exists: %v", err)
		return result
	}

	for _, message := range set.Messages {
		sender, err := self.Whisper.EnsureRecipientExist(message.Sender)
		if err != nil {
			log.Fatalf("%#v", message)
			result.Success = false
			result.Message = fmt.Sprintf("Transfer.TransferGroupMessage: Failed to ensure 'sender' exists: %v", err)
			return result
		}

		_, err = self.Whisper.EnsureMessageExists(
			me,
			sender,
			session,
			message,
		)
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Transfer.TransferGroupMessage: Failed to ensure message exists: %v", err)
			return result
		}
	}

	return result
}

func (self *Transfer) TransferGroupMessages(sets []GroupMessageTransferSet) []GroupMessageTransferResult {
	results := make([]GroupMessageTransferResult, len(sets))
	for i, set := range sets {
		results[i] = self.TransferGroupMessage(set)
	}
	return results
}
