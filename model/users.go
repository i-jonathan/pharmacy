package model

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type User struct {
	baseModel
	UserName string
	Password string
}

func (u *User) HashPassword() error {
	p := passwordConfig{
		time:       3,
		memory:     64 * 1024,
		threads:    2,
		keyLength:  32,
		saltLength: 16,
	}

	salt := make([]byte, p.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	hash := argon2.IDKey([]byte(u.Password), salt, p.time, p.memory, p.threads, p.keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	
    u.Password = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.time, p.threads, b64Salt, b64Hash)
    return nil
}

func (u *User) VerifyPassword(hash string) (bool, error) {
	parts := strings.Split(hash, "$")
	
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash")
	}
	
	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	
	if version != argon2.Version {
		return false, fmt.Errorf("invalid hash version: %d", version)
	}
	
	passConfig := &passwordConfig{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &passConfig.memory, &passConfig.time, &passConfig.threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	passConfig.keyLength = uint32(len(decodedHash))

	comparisonHash := argon2.IDKey([]byte(u.Password), salt, passConfig.time, passConfig.memory, passConfig.threads, passConfig.keyLength)

	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}
