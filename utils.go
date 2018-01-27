package fixr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func uuid() string {
	a := math.Floor(65536 * (1 + rand.Float64()))
	return strconv.FormatInt(int64(a), 16)[1:]
}

func genKey() string {
	s := make([]interface{}, 8)
	for x := range s {
		s[x] = uuid()
	}
	return fmt.Sprintf("%s%s-%s-%s-%s-%s%s%s", s...)
}

type scrapeOutput struct {
	Version string `json:"APP_VERSION"`
}

func updateVersion() error {
	r, err := http.Get("https://fixr.co")
	if err != nil {
		return errors.Wrap(err, "error updating fixr version")
	}
	defer r.Body.Close()
	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "APP_VERSION") {
			s := scanner.Text()
			a, b := strings.Index(s, "{"), strings.Index(s, "}")
			out := new(scrapeOutput)
			if err = json.Unmarshal([]byte(s[a:b+1]), out); err != nil {
				return errors.Wrap(err, "error unmarshalling version")
			}
			fixrVersion = out.Version
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "error scanning version html")
	}
	return nil
}

func SetUserAgent(agent, platform string) {
	userAgent, fixrPlatformVer = agent, platform
}
