package pkce

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"myoidc/pkg/errors"
	"strings"
)

const MethodS256 = "S256"
const MethodPlain = "plain"

const (
	MinLength = 32
	MaxLength = 96
)

type PKCEGenerator interface {
	State() string
	CodeChallengeVerifier() (challenge, method, verifier string)
}

func NewPKCEGenerator(method string, stateLength, challengeLength int) (PKCEGenerator, error) {
	switch method {
	case MethodPlain:
		return NewPlainPKCEGenerator(stateLength, challengeLength)
	default:
		return NewS256PKCEGenerator(stateLength, challengeLength)
	}
}

type S256PKCEGenerator struct {
	StateLength     int
	ChallengeLength int
}

func NewS256PKCEGenerator(stateLength, challengeLength int) (*S256PKCEGenerator, error) {
	if err := validateLength(stateLength, challengeLength); err != nil {
		return nil, nil
	}
	return &S256PKCEGenerator{
		StateLength:     stateLength,
		ChallengeLength: challengeLength,
	}, nil
}

func (v S256PKCEGenerator) State() string {
	return string(random(v.StateLength))
}

func (v S256PKCEGenerator) CodeChallengeVerifier() (challenge, method, verifier string) {
	h := sha256.New()
	vrf := random(v.ChallengeLength)
	h.Write(vrf)
	return v.encode(h.Sum(nil)), "S256", string(vrf)
}

func (v S256PKCEGenerator) encode(msg []byte) string {
	encoded := base64.StdEncoding.EncodeToString(msg)
	encoded = strings.Replace(encoded, "+", "-", -1)
	encoded = strings.Replace(encoded, "/", "_", -1)
	encoded = strings.Replace(encoded, "=", "", -1)
	return encoded
}

type PlainPKCEGenerator struct {
	StateLength     int
	ChallengeLength int
}

func NewPlainPKCEGenerator(stateLength, challengeLength int) (*PlainPKCEGenerator, error) {
	if err := validateLength(stateLength, challengeLength); err != nil {
		return nil, nil
	}
	return &PlainPKCEGenerator{
		StateLength:     stateLength,
		ChallengeLength: challengeLength,
	}, nil
}

func (v PlainPKCEGenerator) State() string {
	return string(random(v.StateLength))
}

func (v PlainPKCEGenerator) CodeChallengeVerifier() (challenge, method, verifier string) {
	value := string(random(v.ChallengeLength))
	return value, "plain", value
}

func validateLength(stateLength, challengeLength int) error {
	if stateLength < MinLength || stateLength > MaxLength {
		return errors.Errorf("invalid state length: %v", stateLength)
	}
	if challengeLength < MinLength || challengeLength > MaxLength {
		return errors.Errorf("invalid challenge length: %v", challengeLength)
	}
	return nil
}

// https://tools.ietf.org/html/rfc7636#section-4.1)
func random(length int) []byte {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	const csLen = byte(len(charset))
	output := make([]byte, 0, length)
	for {
		buf := make([]byte, length)
		if _, err := io.ReadFull(rand.Reader, buf); err != nil {
			return nil
		}
		for _, b := range buf {
			// Avoid bias by using a value range that's a multiple of 62
			if b < (csLen * 4) {
				output = append(output, charset[b%csLen])

				if len(output) == length {
					return output
				}
			}
		}
	}
}
