package ecb

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

const dailyURL = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) FetchDailyRates(ctx context.Context) ([]domain.FXRate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dailyURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ecb request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ecb status: %d", resp.StatusCode)
	}
	return parseECBXML(resp.Body)
}

// ECB XML structure
type envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Cube    struct {
		Cubes []dateCube `xml:"Cube"`
	} `xml:"Cube"`
}

type dateCube struct {
	Time  string     `xml:"time,attr"`
	Rates []rateCube `xml:"Cube"`
}

type rateCube struct {
	Currency string `xml:"currency,attr"`
	Rate     string `xml:"rate,attr"`
}

func parseECBXML(r io.Reader) ([]domain.FXRate, error) {
	var env envelope
	if err := xml.NewDecoder(r).Decode(&env); err != nil {
		return nil, fmt.Errorf("decode ecb xml: %w", err)
	}
	now := time.Now().UTC()
	var out []domain.FXRate
	for _, dc := range env.Cube.Cubes {
		if dc.Time == "" {
			continue
		}
		for _, rc := range dc.Rates {
			rate, err := strconv.ParseFloat(rc.Rate, 64)
			if err != nil {
				continue
			}
			out = append(out, domain.FXRate{
				ID:        uuid.NewString(),
				Date:      dc.Time,
				Base:      "EUR",
				Quote:     rc.Currency,
				Rate:      rate,
				CreatedAt: now,
			})
		}
	}
	return out, nil
}
