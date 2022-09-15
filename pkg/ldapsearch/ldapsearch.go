package ldapsearch

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/go-ldap/ldif"
)

const (
	searchQuery          = "-LLL -z 1 -E pr=1000/noprompt -h ldap.mwam.local -b dc=mwam,dc=local objectClass=user"
	attributeKeyUID      = "uidNumber"
	attributeKeyUsername = "sAMAccountName"
)

type search struct {
	uids map[string]string
}

func Init(testFile string) (*search, error) {
	var output []byte
	var err error
	if testFile == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "ldapsearch", strings.Split(searchQuery, " ")...)
		stderr := ""
		buf := bytes.NewBufferString(stderr)
		cmd.Stderr = buf
		output, err := cmd.Output()
		if err != nil {
			// We need to handle error code 4 specifically because we're requesting too much data and even if ldap is returning it to us, it's still complaining
			if err.(*exec.ExitError).ExitCode() != 4 {
				return nil, fmt.Errorf("error running ldapserarch %v: %v - stderr: %v", err, string(output), buf.String())
			}
			return nil, fmt.Errorf("error running ldapserarch %v: %v - stderr: %v", err, string(output), buf.String())
		}
	} else {
		output, err = ioutil.ReadFile(testFile)
		if err != nil {
			return nil, err
		}
	}
	objects, err := ldif.Parse(string(output))
	if err != nil {
		return nil, err
	}
	u := map[string]string{}
	s := search{
		uids: u,
	}
	for _, entry := range objects.Entries {
		obj := entry.Entry
		uid := ""
		username := ""
		for _, attr := range obj.Attributes {
			if attr.Name == attributeKeyUID {
				uid = strings.Join(attr.Values, "-")
			}
			if attr.Name == attributeKeyUsername {
				username = strings.Join(attr.Values, "-")
				log.Println("Found username: ", username)
			}
		}
		if uid != "" {
			s.uids[uid] = username
		}
	}
	return &s, nil
}

// GetUsername is used to very quickly retried a username from memory
func (s *search) GetUsername(uid string) string {
	user, ok := s.uids[uid]
	if !ok {
		return uid
	}
	return user
}
