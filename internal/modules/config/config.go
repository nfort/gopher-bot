package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/ini.v1"
)

var Config *Configuration

type Configuration struct {
	Tokens map[string]*Token
	Server *ServerConfig
}

func (c *Configuration) Token(instance string) *Token {
	var value *Token
	var ok bool
	if value, ok = c.Tokens[instance]; !ok {
		log.Printf("tokens: %v", c.Tokens)
		panic(fmt.Sprintf("no token for %s", instance))
	}
	return value
}

type Token struct {
	Instance string
	Username string
	Token    string
}

func (t *Token) Git() *http.BasicAuth {
	if t.Username == "" {
		return nil
	}
	return &http.BasicAuth{
		Username: t.Username,
		Password: t.Token,
	}
}

type ServerConfig struct {
	Domain          string `ini:"DOMAIN"`
	Port            int    `ini:"PORT"`
	DebugMode       bool   `ini:"DEBUG_MODE"`
	Secret          string `ini:"SECRET"`
	AllowPush       bool   `ini:"ALLOW_PUSH"`
	AllowPR         bool   `ini:"ALLOW_PR"`
	TLSMode         string `ini:"TLS_MODE"`
	TLSCert         string `ini:"TLS_CERT"`
	TLSPriv         string `ini:"TLS_PRIV"`
	StatusContext   string `ini:"STATUS_CONTEXT"`
	StatusContextPR string `ini:"STATUS_CONTEXT_PR"`
	Skip            string `ini:"SKIP"`
	Owner           string `ini:"OWNER"`
	Repo            string `ini:"REPO"`
}

type DataConfig struct {
	Tmp      bool   `ini:"TMP"`
	Location string `ini:"LOCATION"`
	Keep     bool   `ini:"KEEP"`
}

func InitConfig() error {
	cfg, err := ini.Load("config.ini")
	if os.IsNotExist(err) {
		// use "global" location if file does not not exist
		cfg, err = ini.Load("/etc/gopher-bot/config.ini")
	}
	if err != nil {
		return err
	}

	t := map[string]*Token{}
	for _, k := range cfg.Section("tokens").Keys() {
		tokenParts := strings.Split(k.String(), ":")
		user := ""
		var token string
		if len(tokenParts) < 2 {
			token = tokenParts[0]
		} else {
			user = tokenParts[0]
			token = tokenParts[1]
		}
		t[k.Name()] = &Token{
			Instance: k.Name(),
			Username: user,
			Token:    token,
		}
	}

	s := &ServerConfig{
		Port:            8080,
		DebugMode:       false,
		AllowPush:       true,
		AllowPR:         true,
		StatusContext:   "gopher-bot",
		StatusContextPR: "gopher-bot (PR)",
		Skip:            "[skip gopher-bot]",
	}
	err = cfg.Section("server").MapTo(s)
	if err != nil {
		return err
	}

	Config = &Configuration{
		Tokens: t,
		Server: s,
	}

	return nil
}

func FullURL() string {
	host := Config.Server.Domain
	port := Config.Server.Port
	prot := "http"
	return fmt.Sprintf("%s://%s:%d/", prot, host, port)
}
