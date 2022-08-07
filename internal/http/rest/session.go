package rest

import (
	"encoding/hex"
	"go-sso/internal/storage/fiber_store"
	"math/rand"

	"golang.org/x/crypto/nacl/auth"
)

type Session struct {
	SessionId string `json:"sessionId"`
	UserId    string `json:"userId"`
}

func newSession(userId string) *Session {
	sessionId := make([]byte, 13)
	rand.Read(sessionId)
	sessionIdString := hex.EncodeToString(sessionId)

	return &Session{
		SessionId: sessionIdString,
		UserId:    userId,
	}
}

func (s *Session) saveSession() error {
	err := fiber_store.Store.Storage.Set(s.SessionId, []byte(s.UserId), 0)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) sessionToCookie() string {
	sidBytes, _ := hex.DecodeString(s.SessionId)
	privateKey := getPrivateKeyBytes()

	sig := auth.Sum(sidBytes, &privateKey)
	sigString := hex.EncodeToString(sig[:])
	return s.SessionId + ":" + sigString
}

func getPrivateKeyBytes() [32]byte {
	var privateBytes [32]byte
	privatekey := "kofi-akpor-is-the-big-man"
	copy(privateBytes[:], privatekey)

	return privateBytes
}
