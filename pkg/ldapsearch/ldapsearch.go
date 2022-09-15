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
	searchQuery          = "-LLL -E pr=1000/noprompt -h ldap.mwam.local -b dc=mwam,dc=local (&(objectClass=user)(uidNumber=*)(sAMAccountName=*))"
	attributeKeyUID      = "uidNumber"
	attributeKeyUsername = "sAMAccountName"
)

type Search struct {
	uids map[string]string
}

func Init(testFile string) (*Search, error) {
	var output []byte
	var err error
	if testFile == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "ldapsearch", strings.Split(searchQuery, " ")...)
		stderr := ""
		buf := bytes.NewBufferString(stderr)
		cmd.Stderr = buf
		output, err = cmd.Output()
		if err != nil {
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
	s := Search{
		uids: u,
	}
	log.Printf("Found %v entries! \n", len(objects.Entries))
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
				// log.Println("Found username: ", username)
			}
		}
		if uid != "" {
			s.uids[uid] = username
		}
	}
	// log.Println(s.uids)
	return &s, nil
}

// GetUsername is used to very quickly retried a username from memory
func (s *Search) GetUsername(uid string) string {
	user, ok := s.uids[uid]
	if !ok {
		return uid
	}
	return user
}
