package data

import (
	"time"
)

type User struct {
	Id          string
	ServiceId   string
	Name        string
	FirstName   string
	LastName    string
	FullName    string
	PhoneNumber string
	LastActive  *time.Time
	LastFetched *time.Time
	ProfileKey  []byte
	AvatarUrl   string
	Description string
}

type Group struct {
	Id                 string
	GroupId            []byte
	MasterKey          []byte
	InviteLinkPassword string
	Revision           int
	Name               string
	Members            []GroupMember
	AccessMembers      int
	AccessAttributes   int
	AccessInviteLink   int
	AccessMemberLabels int
	AvatarUrl          string
	Description        string
	AnnouncementOnly   bool
	Terminated         bool
}

type GroupMember struct {
	User             User
	JoinedAtRevision int
	Role             int
}

type MessageType int

const (
	Incoming MessageType = iota
	Outgoing
)

type Message struct {
	Uuid          string
	SentAt        *time.Time
	ReceivedAt    *time.Time
	ExpiresAt     *time.Time
	Sender        User
	Group         *Group
	Type          MessageType
	Body          string
	HasAttachment bool
}

func NewUsers() Users {
	return Users{
		All:         []User{},
		byServiceId: map[string]User{},
		byId:        map[string]User{},
	}
}

type Users struct {
	All         []User
	byServiceId map[string]User
	byId        map[string]User
}

func (users *Users) Add(user User) {
	users.All = append(users.All, user)
	if user.ServiceId != "" {
		users.byServiceId[user.ServiceId] = user
	}
	if user.Id != "" {
		users.byId[user.Id] = user
	}
}

func (users *Users) Remove(user User) {
	var newAll []User
	for _, user := range users.All {
		if user.Id != user.Id {
			newAll = append(newAll, user)
		}
	}
	users.All = newAll
	if user.ServiceId != "" {
		delete(users.byServiceId, user.ServiceId)
	}
	if user.Id != "" {
		delete(users.byId, user.Id)
	}
}

func (users *Users) FindByServiceId(serviceId string) *User {
	for _, user := range users.byServiceId {
		if user.ServiceId == serviceId {
			return &user
		}
	}

	return nil
}

func (users Users) FindById(id string) *User {
	for _, user := range users.byId {
		if user.Id == id {
			return &user
		}
	}

	return nil
}
