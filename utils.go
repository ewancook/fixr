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
	"time"

	"github.com/pkg/errors"
)

var (
	seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func uuid() string {
	segment := math.Floor(65536 * (1 + seededRand.Float64()))
	return strconv.FormatInt(int64(segment), 16)[1:]
}

func genKey() string {
	segments := make([]interface{}, 8)
	for i := range segments {
		segments[i] = uuid()
	}
	return fmt.Sprintf("%s%s-%s-%s-%s-%s%s%s", segments...)
}

type scrapeOutput struct {
	Version string `json:"APP_VERSION"`
}

// UpdateVersion updates the FIXR API version used in the HTTP requests.
func UpdateVersion() error {
	req, err := http.Get(homeURL)
	if err != nil {
		return errors.Wrap(err, "error updating fixr version")
	}
	defer req.Body.Close()
	scanner := bufio.NewScanner(req.Body)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "APP_VERSION") {
			section := scanner.Text()
			start, end := strings.Index(section, "{"), strings.Index(section, "}")
			out := new(scrapeOutput)
			if err = json.Unmarshal([]byte(section[start:end+1]), out); err != nil {
				return errors.Wrap(err, "error unmarshalling version")
			}
			FixrVersion = out.Version
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "error scanning version html")
	}
	return nil
}
