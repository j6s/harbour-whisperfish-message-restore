package whisperfish

import (
	"encoding/hex"
	"fmt"
	"log"
	"maps"
	"slices"
	"time"

	"github.com/j6s/harbour-whisperfish-message-restore/lib"
	"github.com/j6s/harbour-whisperfish-message-restore/lib/data"
	"github.com/jmoiron/sqlx"
)

type WhisperfishRecipient struct {
	Id                    int
	E164                  *string
	Uuid                  string
	Username              *string
	Email                 *string
	IsBlocked             int     `db:"is_blocked"`
	IsAccepted            int     `db:"is_accepted"`
	ProfileKey            []byte  `db:"profile_key"`
	ProfileGivenName      *string `db:"profile_given_name"`
	ProfileFamilyName     *string `db:"profile_family_name"`
	ProfileJoinedName     *string `db:"profile_joined_name"`
	SignalProfileAvatar   *string `db:"signal_profile_avatar"`
	ProfileSharingEnabled int     `db:"profile_sharing_enabled"`
	LastProfileFetch      *string `db:"last_profile_fetch"`
}

func (self *WhisperfishRecipient) ToCols() map[string]interface{} {
	return map[string]interface{}{
		"id":                      self.Id,
		"e164":                    self.E164,
		"uuid":                    self.Uuid,
		"username":                self.Username,
		"email":                   self.Email,
		"is_blocked":              self.IsBlocked,
		"is_accepted":             self.IsAccepted,
		"profile_key":             self.ProfileKey,
		"profile_given_name":      self.ProfileGivenName,
		"profile_family_name":     self.ProfileFamilyName,
		"profile_joined_name":     self.ProfileJoinedName,
		"signal_profile_avatar":   self.SignalProfileAvatar,
		"profile_sharing_enabled": self.ProfileSharingEnabled,
		"last_profile_fetch":      self.LastProfileFetch,
	}
}

type WhisperfishSession struct {
	Id                       int
	DirectMessageRecipientId *int    `db:"direct_message_recipient_id"`
	GroupV2Id                *string `db:"group_v2_id"`
	IsArchived               int     `db:"is_archived"`
	IsPinned                 int     `db:"is_pinned"`
	IsSilent                 int     `db:"is_silent"`
	IsMuted                  int     `db:"is_muted"`
}

func (self *WhisperfishSession) ToCols() map[string]interface{} {
	return map[string]interface{}{
		"id":                          self.Id,
		"direct_message_recipient_id": self.DirectMessageRecipientId,
		"group_v2_id":                 self.GroupV2Id,
		"is_archived":                 self.IsArchived,
		"is_pinned":                   self.IsPinned,
		"is_silent":                   self.IsSilent,
		"is_muted":                    self.IsMuted,
	}
}

type WhisperfishMessage struct {
	Id                int
	SessionId         int `db:"session_id"`
	Text              string
	SenderRecipientId *int    `db:"sender_recipient_id"`
	ReceivedTimestamp *string `db:"received_timestamp"`
	SentTimestamp     *string `db:"sent_timestamp"`
	ServerTimestamp   *string `db:"server_timestamp"`
	IsRead            int     `db:"is_read"`
	IsOutbound        int     `db:"is_outbound"`
	Flags             int     `db:"flags"`
}

func (self *WhisperfishMessage) ToCols() map[string]interface{} {
	return map[string]interface{}{
		"id":                  self.Id,
		"session_id":          self.SessionId,
		"text":                self.Text,
		"sender_recipient_id": self.SenderRecipientId,
		"received_timestamp":  self.ReceivedTimestamp,
		"sent_timestamp":      self.SentTimestamp,
		"server_timestamp":    self.ServerTimestamp,
		"is_read":             self.IsRead,
		"is_outbound":         self.IsOutbound,
		"flags":               self.Flags,
	}
}

type WhisperfishGroup struct {
	Id                 string
	Name               string
	MasterKey          string `db:"master_key"`
	Revision           int
	InviteLinkPassword *string `db:"invite_link_password"`
	Avatar             *string
	Description        *string
	AnnouncementOnly   int `db:"announcement_only"`
}

func (self *WhisperfishGroup) ToCols() map[string]interface{} {
	return map[string]interface{}{
		"id":                   self.Id,
		"name":                 self.Name,
		"master_key":           self.MasterKey,
		"revision":             self.Revision,
		"invite_link_password": self.InviteLinkPassword,
		"description":          self.Description,
		"announcement_only":    self.AnnouncementOnly,
	}
}

type WhisperfishGroupMember struct {
	GroupV2Id        string `db:"group_v2_id"`
	RecipientId      int    `db:"recipient_id"`
	JoinedAtRevision int    `db:"joined_at_revision"`
	Role             int    `db:"role"`
}

func (self *WhisperfishGroupMember) ToCols() map[string]interface{} {
	return map[string]interface{}{
		"group_v2_id":        self.GroupV2Id,
		"recipient_id":       self.RecipientId,
		"joined_at_revision": self.JoinedAtRevision,
		"role":               self.Role,
	}
}

func CreateWhisperfish(file string) (*Whisperfish, error) {
	connection, err := sqlx.Connect("sqlite", fmt.Sprintf("file:%s", file))
	if err != nil {
		return nil, fmt.Errorf("Cannot open SQLite connection to %s: %v", file, err)
	}

	return &Whisperfish{
		db: connection,
	}, nil
}

type Whisperfish struct {
	db *sqlx.DB
}

func (self *Whisperfish) EnsureRecipientExist(user data.User) (WhisperfishRecipient, error) {
	if user.ServiceId == "" {
		return WhisperfishRecipient{}, fmt.Errorf("Whisperfish.EnsureRecipientExists: Cannot create or update user without serviceId")
	}

	result, err := self.db.Queryx("SELECT id, e164, uuid, username, email, is_blocked, profile_key, profile_given_name, profile_family_name, profile_joined_name, signal_profile_avatar, profile_sharing_enabled, last_profile_fetch FROM recipients WHERE uuid=? OR e164=?", user.ServiceId, user.PhoneNumber)
	if err != nil {
		return WhisperfishRecipient{}, fmt.Errorf("Whisperfish.EnsureRecipientExists: Cannot query existing recipients: %v", err)
	}
	defer result.Close()

	isExistingUser := result.Next()
	var recipient WhisperfishRecipient
	if isExistingUser {
		err = result.StructScan(&recipient)
		if err != nil {
			return recipient, fmt.Errorf("Whisperfish.EnsureRecipientExists: Cannot map existing recipients: %v", err)
		}

		return recipient, nil
	}
	maxId, err := lib.GetMaxId(self.db, "recipients", "id")
	if err != nil {
		return recipient, fmt.Errorf("Whisperfish.EnsureRecipientExists: Cannot determine next recipient id: %v", err)
	}

	recipient = WhisperfishRecipient{
		Id:                    maxId + 1,
		E164:                  &user.PhoneNumber,
		Uuid:                  user.ServiceId,
		Username:              nil,
		Email:                 nil,
		IsBlocked:             0,
		IsAccepted:            1,
		ProfileKey:            user.ProfileKey,
		ProfileGivenName:      &user.Name,
		ProfileFamilyName:     &user.LastName,
		ProfileJoinedName:     &user.FullName,
		SignalProfileAvatar:   &user.AvatarUrl,
		ProfileSharingEnabled: 0,
		LastProfileFetch:      formatDateForDatabase(user.LastFetched),
	}

	_, err = lib.BuildInsertQuery(self.db, "recipients", recipient.ToCols())()
	if err != nil {
		return recipient, fmt.Errorf("Whisperfish.EnsureRecipientExists: Error while executing UPDATE/INSERT query: %v", err)
	}

	return recipient, nil
}

func (self *Whisperfish) EnsurePrivateMessageSessionExists(recipient WhisperfishRecipient) (WhisperfishSession, error) {
	cols := slices.Collect(maps.Keys((&WhisperfishSession{}).ToCols()))
	result, err := lib.BuildSelectStatement(self.db, "sessions", cols, map[string]interface{}{
		"direct_message_recipient_id": recipient.Id,
	})()
	if err != nil {
		return WhisperfishSession{}, fmt.Errorf("Whisperfish.EnsurePrivateMessageSessionExists: Cannot check if session exists: %v", err)
	}
	defer result.Close()

	exists := result.Next()
	var session WhisperfishSession
	if exists {
		err = result.StructScan(&session)
		if err != nil {
			return session, fmt.Errorf("Whisperfish.EnsurePrivateMessageSessionExists: Cannot map existing session: %v", err)
		}
		return session, nil
	}

	maxId, err := lib.GetMaxId(self.db, "sessions", "id")
	if err != nil {
		return session, fmt.Errorf("Whisperfish.EnsurePrivateMessageSessionExists: Cannot determine next session id: %v", err)
	}

	session = WhisperfishSession{
		Id:                       maxId + 1,
		DirectMessageRecipientId: &recipient.Id,
		GroupV2Id:                nil,
		IsArchived:               0,
		IsMuted:                  0,
		IsSilent:                 0,
		IsPinned:                 0,
	}

	_, err = lib.BuildInsertQuery(self.db, "sessions", session.ToCols())()
	if err != nil {
		return session, fmt.Errorf("Whisperfish.EnsurePrivateMessageSessionExists: Cannot insert session: %v", err)
	}

	return session, nil
}

func (self *Whisperfish) EnsureGroupExists(group data.Group) (WhisperfishGroup, error) {
	var whisperfishGroup WhisperfishGroup

	whisperfishGroupId := hex.EncodeToString(group.GroupId)
	cols := slices.Collect(maps.Keys((&WhisperfishGroup{}).ToCols()))
	result, err := lib.BuildSelectStatement(self.db, "group_v2s", cols, map[string]interface{}{
		"id": whisperfishGroupId,
	})()
	if err != nil {
		return whisperfishGroup, fmt.Errorf("Whisperfish.EnsureGroupExists: Cannot check if group exists: %v", err)
	}
	defer result.Close()

	exists := result.Next()
	if exists {
		err = result.StructScan(&whisperfishGroup)
		if err != nil {
			return whisperfishGroup, fmt.Errorf("Whisperfish.EnsureGroupExists: Cannot map existing group: %v", err)
		}

		return whisperfishGroup, nil
	}

	announcementOnly := 0
	if group.AnnouncementOnly {
		announcementOnly = 1
	}
	whisperfishGroup = WhisperfishGroup{
		Id:                 whisperfishGroupId,
		Name:               group.Name,
		MasterKey:          hex.EncodeToString(group.MasterKey),
		Revision:           group.Revision,
		InviteLinkPassword: &group.InviteLinkPassword,
		Avatar:             &group.AvatarUrl,
		Description:        &group.Description,
		AnnouncementOnly:   announcementOnly,
	}

	_, err = lib.BuildInsertQuery(self.db, "group_v2s", whisperfishGroup.ToCols())()
	if err != nil {
		return whisperfishGroup, fmt.Errorf("Whisperfish.EnsureGroupExists: Cannot insert group: %v", err)
	}

	return whisperfishGroup, nil
}

func (self *Whisperfish) EnsureGroupMembersExist(group data.Group, whisperfishGroup WhisperfishGroup) error {
	for _, member := range group.Members {
		_, err := self.EnsureGroupMemberExists(whisperfishGroup, member)
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *Whisperfish) EnsureGroupMemberExists(whisperfishGroup WhisperfishGroup, member data.GroupMember) (WhisperfishGroupMember, error) {
	var whisperfishGroupMember WhisperfishGroupMember

	recipient, err := self.EnsureRecipientExist(member.User)
	if err != nil {
		return whisperfishGroupMember, fmt.Errorf("Whisperfish.EnsureGroupMembersExist: Cannot ensure recipient exists: %v", err)
	}

	cols := slices.Collect(maps.Keys((&WhisperfishGroupMember{}).ToCols()))
	result, err := lib.BuildSelectStatement(self.db, "group_v2_members", cols, map[string]interface{}{
		"group_v2_id":  whisperfishGroup.Id,
		"recipient_id": recipient.Id,
	})()
	if err != nil {
		return whisperfishGroupMember, fmt.Errorf("Whisperfish.EnsureGroupMembersExist: Cannot query existing group: %v", err)
	}
	defer result.Close()

	exists := result.Next()
	if exists {
		err = result.StructScan(&whisperfishGroupMember)
		if err != nil {
			return whisperfishGroupMember, fmt.Errorf("Whisperfish.EnsureGroupMembersExist: Cannot map existing group: %v", err)
		}

		return whisperfishGroupMember, nil
	}

	whisperfishGroupMember = WhisperfishGroupMember{
		GroupV2Id:        whisperfishGroup.Id,
		RecipientId:      recipient.Id,
		JoinedAtRevision: member.JoinedAtRevision,
		Role:             0,
	}

	_, err = lib.BuildInsertQuery(self.db, "group_v2_members", whisperfishGroupMember.ToCols())()
	if err != nil {
		return whisperfishGroupMember, fmt.Errorf("Whisperfish.EnsureGroupMembersExist: Cannot insert group: %v", err)
	}

	return whisperfishGroupMember, nil
}

func (self *Whisperfish) EnsureGroupMessageSessionExists(group WhisperfishGroup) (WhisperfishSession, error) {
	cols := slices.Collect(maps.Keys((&WhisperfishSession{}).ToCols()))
	result, err := lib.BuildSelectStatement(self.db, "sessions", cols, map[string]interface{}{
		"group_v2_id": group.Id,
	})()
	if err != nil {
		return WhisperfishSession{}, fmt.Errorf("Whisperfish.EnsureGroupMessageSessionExists: Cannot check if session exists: %v", err)
	}
	defer result.Close()

	exists := result.Next()
	var session WhisperfishSession
	if exists {
		err = result.StructScan(&session)
		if err != nil {
			return session, fmt.Errorf("Whisperfish.EnsureGroupMessageSessionExists: Cannot map existing session: %v", err)
		}
		return session, nil
	}

	maxId, err := lib.GetMaxId(self.db, "sessions", "id")
	if err != nil {
		return session, fmt.Errorf("Whisperfish.EnsureGroupMessageSessionExists: Cannot determine next session id: %v", err)
	}

	session = WhisperfishSession{
		Id:                       maxId + 1,
		DirectMessageRecipientId: nil,
		GroupV2Id:                &group.Id,
		IsArchived:               0,
		IsMuted:                  0,
		IsSilent:                 0,
		IsPinned:                 0,
	}

	_, err = lib.BuildInsertQuery(self.db, "sessions", session.ToCols())()
	if err != nil {
		return session, fmt.Errorf("Whisperfish.EnsureGroupMessageSessionExists: Cannot insert session: %v", err)
	}

	return session, nil
}

func (self *Whisperfish) EnsureMessageExists(
	me WhisperfishRecipient,
	recipient WhisperfishRecipient,
	session WhisperfishSession,
	message data.Message,
) (WhisperfishMessage, error) {

	var sender WhisperfishRecipient
	var isOutbound int
	if message.Type == data.Outgoing {
		sender = me
		isOutbound = 1
	} else {
		sender = recipient
		isOutbound = 0
	}

	body := message.Body
	if message.HasAttachment {
		body = fmt.Sprintf("%s\n\n[Attachment not imported]", body)
	}

	whisperfishMessage := WhisperfishMessage{
		SessionId:         session.Id,
		Text:              body,
		SenderRecipientId: &sender.Id,
		ReceivedTimestamp: formatDateForDatabase(message.ReceivedAt),
		SentTimestamp:     formatDateForDatabase(message.SentAt),
		ServerTimestamp:   formatDateForDatabase(message.SentAt),
		IsRead:            1,
		IsOutbound:        isOutbound,
		Flags:             0,
	}

	cols := slices.Collect(maps.Keys((&WhisperfishMessage{}).ToCols()))
	result, err := lib.BuildSelectStatement(self.db, "messages", cols, map[string]interface{}{
		"session_id":     whisperfishMessage.SessionId,
		"text":           whisperfishMessage.Text,
		"sent_timestamp": whisperfishMessage.SentTimestamp,
	})()
	if err != nil {
		return whisperfishMessage, fmt.Errorf("Whisperfish.EnsureMessageExists: Cannot check for existing message: %v", err)
	}
	defer result.Close()

	exists := result.Next()
	if exists {
		err = result.StructScan(&whisperfishMessage)
		if err != nil {
			log.Fatal(err)
			return whisperfishMessage, fmt.Errorf("Whisperfish.EnsureMessageExists: Cannot map existing message: %v", err)
		}

		return whisperfishMessage, nil
	}

	maxId, err := lib.GetMaxId(self.db, "messages", "id")
	if err != nil {
		return whisperfishMessage, fmt.Errorf("Whisperfish.EnsureMessageExists: Cannot determine next message id: %v", err)
	}

	whisperfishMessage.Id = maxId + 1
	if err != nil {
		return whisperfishMessage, fmt.Errorf("Whisperfish.EnsureMessageExists: Cannot count existing messages: %v", err)
	}

	_, err = lib.BuildInsertQuery(self.db, "messages", whisperfishMessage.ToCols())()
	if err != nil {
		return whisperfishMessage, fmt.Errorf("Whisperfish.EnsureMessageExists: Cannot INSERT new message: %v", err)
	}

	return whisperfishMessage, nil
}

func formatDateForDatabase(date *time.Time) *string {
	if date == nil {
		return nil
	}

	formatted := date.Format("2006-01-02 15:04:05.000000000")

	return &formatted
}
