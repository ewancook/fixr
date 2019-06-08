package fixr

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
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

func buildURLValues(values payload) (url.Values, error) {
	pl := url.Values{}
	for key, value := range values {
		valueStr, ok := value.(string)
		if !ok {
			return nil, errors.New("failed to build payload")
		}
		pl.Set(key, valueStr)
	}
	return pl, nil
}

func jsonifyPayload(kval payload) (*bytes.Buffer, error) {
	data := new(bytes.Buffer)
	if err := json.NewEncoder(data).Encode(kval); err != nil {
		return nil, errors.Wrap(err, "error jsonifying payload")
	}
	return data, nil
}

type scrapeOutput struct {
	Version string `json:"APP_VERSION"`
}

func unmarshalOutput(section string) (*scrapeOutput, error) {
	output := new(scrapeOutput)
	start, end := strings.Index(section, "{"), strings.Index(section, "}")
	if err := json.Unmarshal([]byte(section[start:end+1]), output); err != nil {
		return nil, errors.Wrap(err, "unmarshalling failure")
	}
	return output, nil
}

// UpdateVersion updates the FIXR API version used in the HTTP requests.
func UpdateVersion() error {
	req, err := http.Get(homeURL)
	if err != nil {
		return errors.Wrap(err, "failed to update fixr.FixrVersion")
	}
	defer req.Body.Close()
	scanner := bufio.NewScanner(req.Body)
	for scanner.Scan() {
		section := scanner.Text()
		if !strings.Contains(section, "APP_VERSION") {
			continue
		}
		out, err := unmarshalOutput(section)
		if len(out.Version) > 0 {
			FixrVersion = out.Version
		}
		return err
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to scan HTML for %s", homeURL))
	}
	return errors.New("search for APP_VERSION failed")
}
