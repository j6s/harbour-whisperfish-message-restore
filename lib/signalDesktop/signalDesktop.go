package signalDesktop

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/j6s/harbour-whisperfish-message-restore/lib"
	"github.com/j6s/harbour-whisperfish-message-restore/lib/data"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type SignalDesktopConversationRow struct {
	Id                   string
	Json                 string
	ActiveAt             *int `db:"active_at"`
	Type                 string
	Members              *string
	Name                 *string
	ProfileName          *string `db:"profileName"`
	ProfileFamilyName    *string `db:"profileFamilyName"`
	ProfileFullName      *string `db:"profileFullName"`
	E164                 *string `db:"e164"`
	ServiceId            *string `db:"serviceId"`
	GroupId              *string `db:"groupId"`
	ProfileLastFetchedAt *int    `db:"profileLastFetchedAt"`
}

type SignalPrivateConversationJson struct {
	Id                            string        `json:"id"`
	Type                          string        `json:"type"`
	Version                       int           `json:"version"`
	ServiceId                     string        `json:"serviceId"`
	Pni                           string        `json:"pni"`
	E164                          string        `json:"e164"`
	ProfileKey                    string        `json:"profileKey"`
	AccessKey                     string        `json:"accessKey"`
	SealedSender                  int           `json:"sealedSender"`
	ProfileSharing                bool          `json:"profileSharing"`
	ProfileName                   string        `json:"profileName"`
	ProfileFamilyName             string        `json:"profileFamilyName"`
	SystemGivenName               string        `json:"systemGivenName"`
	SystemFamilyName              string        `json:"systemFamilyName"`
	HideStory                     bool          `json:"hideStory"`
	Note                          string        `json:"note"`
	Color                         string        `json:"color"`
	ColorFromPrimary              int           `json:"colorFromPrimary"`
	Verified                      int           `json:"verified"`
	IsArchived                    bool          `json:"isArchived"`
	IsPinned                      bool          `json:"isPinned"`
	MarkedUnread                  bool          `json:"markedUnread"`
	DontNotifyForMentionsIfMuted  bool          `json:"dontNotifyForMentionsIfMuted"`
	ConversationColor             string        `json:"conversationColor"`
	ActiveAt                      int64         `json:"active_at"`
	MessageCount                  int           `json:"messageCount"`
	SentMessageCount              int           `json:"sentMessageCount"`
	ExpireTimerVersion            int           `json:"expireTimerVersion"`
	LastMessage                   string        `json:"lastMessage"`
	LastMessageBodyRanges         []interface{} `json:"lastMessageBodyRanges"`
	LastMessagePrefix             string        `json:"lastMessagePrefix"`
	LastMessageAuthor             string        `json:"lastMessageAuthor"`
	LastMessageStatus             string        `json:"lastMessageStatus"`
	LastMessageReceivedAt         int64         `json:"lastMessageReceivedAt"`
	LastMessageReceivedAtMs       int64         `json:"lastMessageReceivedAtMs"`
	Timestamp                     int64         `json:"timestamp"`
	LastMessageDeletedForEveryone bool          `json:"lastMessageDeletedForEveryone"`
	LastMessageAuthorAci          string        `json:"lastMessageAuthorAci"`
	About                         string        `json:"about"`
	AboutEmoji                    string        `json:"aboutEmoji"`
	SharingPhoneNumber            bool          `json:"sharingPhoneNumber"`
	Capabilities                  struct {
		ProfilesV2                bool `json:"profiles_v2"`
		AttachmentBackfill        bool `json:"attachmentBackfill"`
		Spqr                      bool `json:"spqr"`
		UsernameChangeSyncMessage bool `json:"usernameChangeSyncMessage"`
	} `json:"capabilities"`
	ProfileKeyCredential           string `json:"profileKeyCredential"`
	ProfileKeyCredentialExpiration int64  `json:"profileKeyCredentialExpiration"`
	ProfileAvatar                  struct {
		Url           string `json:"url"`
		Version       int    `json:"version"`
		PlaintextHash string `json:"plaintextHash"`
		Size          int    `json:"size"`
		Path          string `json:"path"`
		LocalKey      string `json:"localKey"`
	} `json:"profileAvatar"`
	LastProfile struct {
		ProfileKey        string `json:"profileKey"`
		ProfileKeyVersion string `json:"profileKeyVersion"`
	} `json:"lastProfile"`
	Name          string `json:"name"`
	InboxPosition int    `json:"inbox_position"`
	Avatar        struct {
		Unknown       []interface{} `json:"$unknown"`
		ContentType   string        `json:"contentType"`
		Length        int           `json:"length"`
		Version       int           `json:"version"`
		PlaintextHash string        `json:"plaintextHash"`
		Size          int           `json:"size"`
		Path          string        `json:"path"`
		LocalKey      string        `json:"localKey"`
		Hash          string        `json:"hash"`
	} `json:"avatar"`
	MessageRequestResponseType int           `json:"messageRequestResponseType"`
	StorageUnknownFields       string        `json:"storageUnknownFields"`
	StorageID                  string        `json:"storageID"`
	StorageVersion             int           `json:"storageVersion"`
	NeedsStorageServiceSync    bool          `json:"needsStorageServiceSync"`
	MuteExpiresAt              int           `json:"muteExpiresAt"`
	UnreadCount                int           `json:"unreadCount"`
	UnreadMentionsCount        int           `json:"unreadMentionsCount"`
	Draft                      string        `json:"draft"`
	DraftBodyRanges            []interface{} `json:"draftBodyRanges"`
	DraftChanged               bool          `json:"draftChanged"`
	DraftIsViewOnce            bool          `json:"draftIsViewOnce"`
}

type SignalGroupConversationJson struct {
	Id                         string `json:"id"`
	Type                       string `json:"type"`
	Version                    int    `json:"version"`
	GroupVersion               int    `json:"groupVersion"`
	MasterKey                  string `json:"masterKey"`
	GroupId                    string `json:"groupId"`
	SecretParams               string `json:"secretParams"`
	PublicParams               string `json:"publicParams"`
	ProfileSharing             bool   `json:"profileSharing"`
	MessageRequestResponseType int    `json:"messageRequestResponseType"`
	HideStory                  bool   `json:"hideStory"`
	Avatar                     struct {
		Url           string `json:"url"`
		Version       int    `json:"version"`
		PlaintextHash string `json:"plaintextHash"`
		Size          int    `json:"size"`
		Path          string `json:"path"`
		LocalKey      string `json:"localKey"`
		Hash          string `json:"hash"`
	} `json:"avatar"`
	RemoteAvatarUrl  string `json:"remoteAvatarUrl"`
	Color            string `json:"color"`
	ColorFromPrimary int    `json:"colorFromPrimary"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	ExpireTimer      int    `json:"expireTimer"`
	AccessControl    struct {
		Members           int `json:"members"`
		Attributes        int `json:"attributes"`
		AddFromInviteLink int `json:"addFromInviteLink"`
		MemberLabel       int `json:"memberLabel"`
	} `json:"accessControl"`
	MembersV2 []struct {
		Role            int    `json:"role"`
		JoinedAtVersion int    `json:"joinedAtVersion"`
		Aci             string `json:"aci"`
	} `json:"membersV2"`
	PendingMembersV2       []interface{} `json:"pendingMembersV2"`
	PendingAdminApprovalV2 []interface{} `json:"pendingAdminApprovalV2"`
	BannedMembersV2        []struct {
		ServiceId string `json:"serviceId"`
		Timestamp int64  `json:"timestamp"`
	} `json:"bannedMembersV2"`
	Revision                      int         `json:"revision"`
	GroupInviteLinkPassword       string      `json:"groupInviteLinkPassword"`
	AnnouncementsOnly             bool        `json:"announcementsOnly"`
	Terminated                    bool        `json:"terminated"`
	IsArchived                    bool        `json:"isArchived"`
	IsPinned                      bool        `json:"isPinned"`
	MarkedUnread                  bool        `json:"markedUnread"`
	DontNotifyForMentionsIfMuted  bool        `json:"dontNotifyForMentionsIfMuted"`
	AutoBubbleColor               bool        `json:"autoBubbleColor"`
	ActiveAt                      int64       `json:"active_at"`
	ExpireTimerVersion            int         `json:"expireTimerVersion"`
	SealedSender                  int         `json:"sealedSender"`
	LastMessage                   string      `json:"lastMessage"`
	LastMessageReceivedAt         int64       `json:"lastMessageReceivedAt"`
	Timestamp                     int64       `json:"timestamp"`
	LastMessageDeletedForEveryone bool        `json:"lastMessageDeletedForEveryone"`
	LastMessageAuthorAci          interface{} `json:"lastMessageAuthorAci"`
	StorageID                     string      `json:"storageID"`
	StorageVersion                int         `json:"storageVersion"`
	StorySendMode                 string      `json:"storySendMode"`
	NeedsStorageServiceSync       bool        `json:"needsStorageServiceSync"`
	MuteExpiresAt                 int         `json:"muteExpiresAt"`
	StorageUnknownFields          string      `json:"storageUnknownFields"`
	Left                          bool        `json:"left"`
}

type SignalDesktopMessage struct {
	Id              string
	Json            string
	SentAt          *int   `db:"sent_at"`
	ConversationId  string `db:"conversationId"`
	ReceivedAt      *int   `db:"received_at"`
	Type            string
	Body            *string
	SourceServiceId *string `db:"sourceServiceId"`
	HasAttachments  int     `db:"hasAttachments"`
}

func (self *SignalDesktopMessage) ToCols() map[string]interface{} {
	return map[string]interface{}{
		"id":              self.Id,
		"json":            self.Json,
		"sent_at":         self.SentAt,
		"conversationId":  self.ConversationId,
		"received_at":     self.ReceivedAt,
		"type":            self.Type,
		"body":            self.Body,
		"sourceServiceId": self.SourceServiceId,
	}
}

func CreateSignalDesktop(file string) (*SignalDesktop, error) {
	connection, err := sqlx.Connect("sqlite", fmt.Sprintf("file:%s", file))
	if err != nil {
		return nil, fmt.Errorf("CreateSignalDesktop: Cannot open SQLite connection to %s: %v", file, err)
	}

	return &SignalDesktop{
		db: connection,
	}, nil
}

type SignalDesktop struct {
	db *sqlx.DB
}

func (self *SignalDesktop) FindMe(users data.Users) (data.User, error) {
	result, err := self.db.Queryx("SELECT json FROM items WHERE id=\"uuid_id\"")
	if err != nil {
		return data.User{}, fmt.Errorf("SignalDesktop.FindMe: Cannot execute config query: %v", err)
	}
	defer result.Close()

	if !result.Next() {
		return data.User{}, fmt.Errorf("SignalDesktop.FindMe: Config query has no result: %v", result.Err())
	}

	row := struct{ Json string }{}
	err = result.StructScan(&row)
	if err != nil {
		return data.User{}, fmt.Errorf("SignalDesktop.FindMe: Cannot map config query result: %v", err)
	}

	jsonContent := struct{ Value string }{}
	err = json.Unmarshal([]byte(row.Json), &jsonContent)
	if err != nil {
		return data.User{}, fmt.Errorf("SignalDesktop.FindMe: Cannot parse Config JSON content: %v", err)
	}

	id := strings.Split(jsonContent.Value, ".")[0]
	me := users.FindByServiceId(id)
	if me == nil {
		return data.User{}, fmt.Errorf("SignalDesktop.FindMe: User with serviceId %s does not exist", id)
	}

	return *me, nil
}

func (self *SignalDesktop) GetUsers() (data.Users, error) {
	users := data.NewUsers()
	conversations, err := self.fetchConversations("private")
	if err != nil {
		return users, err
	}

	for _, row := range conversations {
		user, err := self.mapUser(row)
		if err != nil {
			return users, err
		}

		users.Add(user)
	}

	return users, nil
}

func (self *SignalDesktop) mapUser(row SignalDesktopConversationRow) (data.User, error) {
	var jsonContent SignalPrivateConversationJson
	err := json.Unmarshal([]byte(row.Json), &jsonContent)
	if err != nil {
		return data.User{}, fmt.Errorf("SignalDesktop.mapUser: Cannot unmarshal %s: %v", row.Json, err)
	}

	profileKey, err := base64.StdEncoding.DecodeString(jsonContent.ProfileKey)
	if err != nil {
		return data.User{}, fmt.Errorf("SignalDesktop.mapUser: Cannot decode %s: %v", jsonContent.ProfileKey, err)
	}

	return data.User{
		Id:          jsonContent.Id,
		ServiceId:   parseString(row.ServiceId),
		Name:        parseString(row.Name),
		FirstName:   parseString(row.ProfileName),
		LastName:    parseString(row.ProfileFamilyName),
		FullName:    parseString(row.ProfileFullName),
		PhoneNumber: parseString(row.E164),
		LastActive:  parseTimestamp(row.ProfileLastFetchedAt),
		LastFetched: parseTimestamp(row.ProfileLastFetchedAt),
		ProfileKey:  profileKey,
		AvatarUrl:   jsonContent.ProfileAvatar.Url,
		Description: "",
	}, nil
}

func (self *SignalDesktop) GetGroups(users data.Users) ([]data.Group, error) {
	conversations, err := self.fetchConversations("group")
	if err != nil {
		return nil, err
	}

	var groups []data.Group
	for _, row := range conversations {
		var jsonContent SignalGroupConversationJson
		err = json.Unmarshal([]byte(row.Json), &jsonContent)
		if err != nil {
			return groups, fmt.Errorf("SignalDesktop.GetGroups: Cannot unmarshal %s: %v", row.Json, err)
		}

		var members []data.GroupMember
		for _, groupMember := range jsonContent.MembersV2 {
			user := users.FindByServiceId(groupMember.Aci)
			if user == nil {
				log.Printf("[WARN] Cannot resolve group member %#v of group %#v", groupMember, jsonContent)
				continue
			}

			members = append(members, data.GroupMember{
				Role:             groupMember.Role,
				JoinedAtRevision: groupMember.JoinedAtVersion,
				User:             *user,
			})
		}

		masterKey, err := base64.StdEncoding.DecodeString(jsonContent.MasterKey)
		if err != nil {
			return groups, fmt.Errorf("SignalDesktop.GetGroups: Cannot decode %s: %v", jsonContent.MasterKey, err)
		}

		groupId, err := base64.StdEncoding.DecodeString(jsonContent.GroupId)
		if err != nil {
			return groups, fmt.Errorf("SignalDesktop.GetGroups: Cannot decode %s: %v", jsonContent.GroupId, err)
		}

		groups = append(groups, data.Group{
			Id:                 row.Id,
			GroupId:            groupId,
			MasterKey:          masterKey,
			InviteLinkPassword: jsonContent.GroupInviteLinkPassword,
			Revision:           jsonContent.Revision,
			Name:               parseString(row.Name),
			Members:            members,
			AccessMembers:      jsonContent.AccessControl.Members,
			AccessAttributes:   jsonContent.AccessControl.Attributes,
			AccessInviteLink:   jsonContent.AccessControl.MemberLabel,
			AvatarUrl:          jsonContent.RemoteAvatarUrl,
			Description:        jsonContent.Description,
			AnnouncementOnly:   jsonContent.AnnouncementsOnly,
			Terminated:         jsonContent.Terminated,
		})
	}

	return groups, nil
}

func (self *SignalDesktop) GetPrivateMessages(targetUser data.User, me data.User) ([]data.Message, error) {
	var messages []data.Message
	result, err := self.fetchMessages(targetUser.Id)
	if err != nil {
		return messages, err
	}

	for _, row := range result {
		var sender data.User
		var messageType data.MessageType
		if row.Type == "incoming" {
			sender = targetUser
			messageType = data.Incoming
		} else if row.Type == "outgoing" {
			sender = me
			messageType = data.Outgoing
		} else {
			continue
		}

		messages = append(messages, data.Message{
			Uuid:          row.Id,
			SentAt:        parseTimestamp(row.SentAt),
			ReceivedAt:    parseTimestamp(row.ReceivedAt),
			ExpiresAt:     nil,
			Sender:        sender,
			Group:         nil,
			Type:          messageType,
			Body:          parseString(row.Body),
			HasAttachment: row.HasAttachments == 1,
		})
	}

	return messages, nil
}

func (self *SignalDesktop) GetGroupMessages(group data.Group, me data.User) ([]data.Message, error) {
	var messages []data.Message

	membersById := map[string]data.User{}
	for _, member := range group.Members {
		membersById[member.User.ServiceId] = member.User
	}

	result, err := self.fetchMessages(group.Id)
	if err != nil {
		return messages, err
	}

	for _, row := range result {
		var sender data.User
		var messageType data.MessageType
		if row.Type == "incoming" {
			sender = membersById[*row.SourceServiceId]
			messageType = data.Incoming
		} else if row.Type == "outgoing" {
			sender = me
			messageType = data.Outgoing
		} else {
			continue
		}

		messages = append(messages, data.Message{
			Uuid:          row.Id,
			SentAt:        parseTimestamp(row.SentAt),
			ReceivedAt:    parseTimestamp(row.ReceivedAt),
			ExpiresAt:     nil,
			Sender:        sender,
			Group:         &group,
			Type:          messageType,
			Body:          parseString(row.Body),
			HasAttachment: row.HasAttachments == 1,
		})
	}

	return messages, nil
}

func (self *SignalDesktop) fetchMessages(conversationId string) ([]SignalDesktopMessage, error) {
	var messages []SignalDesktopMessage
	result, err := lib.BuildSelectStatement(
		self.db,
		"messages",
		slices.Collect(maps.Keys((&SignalDesktopMessage{}).ToCols())),
		map[string]interface{}{"conversationId": conversationId},
	)()
	defer result.Close()

	if err != nil {
		return messages, fmt.Errorf("SignalDesktop.GetPrivateMessages: Cannot query Messages: %v", err)
	}

	defer result.Close()
	for result.Next() {
		var row SignalDesktopMessage
		err = result.StructScan(&row)
		if err != nil {
			return messages, fmt.Errorf("SignalDesktop.GetPrivateMessages: Cannot map Messages: %v", err)
		}

		messages = append(messages, row)
	}

	return messages, nil
}

func (self *SignalDesktop) fetchConversations(conversationType string) ([]SignalDesktopConversationRow, error) {
	// TODO: USE BUILD SELECT STATEMENT
	var rows []SignalDesktopConversationRow
	converations, err := self.db.Queryx("SELECT id, json, active_at, type, members, name, profileName, profileFamilyName, profileFullName, e164, serviceId, groupId, profileLastFetchedAt FROM conversations WHERE type=$1", conversationType)
	if err != nil {
		return rows, fmt.Errorf("SignalDesktop.fetchConversation: Cannot query conversations: %v", err)
	}
	defer converations.Close()

	defer converations.Close()
	for converations.Next() {
		var row SignalDesktopConversationRow
		err = converations.StructScan(&row)
		if err != nil {
			return rows, fmt.Errorf("SignalDesktop.GetPrivateMessages: Cannot map conversations: %v", err)
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func parseTimestamp(timestamp *int) *time.Time {
	if timestamp == nil {
		return nil
	}

	parsed := time.Unix(int64(*timestamp)/1000, 0)

	return &parsed
}

func parseString(string *string) string {
	if string == nil {
		return ""
	}

	return *string
}
