package maxmind

import (
	"errors"
	"net"
	"sync"

	"github.com/oschwald/maxminddb-golang"
)

var (
	ErrNoDB      = errors.New("no database loaded")
	ErrNotFound  = errors.New("record not found")
	ErrInvalidIP = errors.New("invalid ip address")
)

type Continent struct {
	Code  string            `maxminddb:"code"`
	Names map[string]string `maxminddb:"names"`
}

type Country struct {
	ISOCode string            `maxminddb:"iso_code"`
	Names   map[string]string `maxminddb:"names"`
}

type City struct {
	Names map[string]string `maxminddb:"names"`
}

type Postal struct {
	Code string `maxminddb:"code"`
}

type Location struct {
	Latitude       float64 `maxminddb:"latitude"`
	Longitude      float64 `maxminddb:"longitude"`
	AccuracyRadius int     `maxminddb:"accuracy_radius"`
	TimeZone       string  `maxminddb:"time_zone"`
}

type Traits struct {
	IsAnonymousProxy    bool `maxminddb:"is_anonymous_proxy"`
	IsSatelliteProvider bool `maxminddb:"is_satellite_provider"`
}

type CityRecord struct {
	Continent    Continent `maxminddb:"continent"`
	Country      Country   `maxminddb:"country"`
	Subdivisions []Country `maxminddb:"subdivisions"`
	City         City      `maxminddb:"city"`
	Postal       Postal    `maxminddb:"postal"`
	Location     Location  `maxminddb:"location"`
	Traits       Traits    `maxminddb:"traits"`
}

type CountryRecord struct {
	Continent Continent `maxminddb:"continent"`
	Country   Country   `maxminddb:"country"`
}

type ASNRecord struct {
	Number int    `maxminddb:"autonomous_system_number"`
	Org    string `maxminddb:"autonomous_system_organization"`
}

type CityResult struct {
	CityRecord
	Network *net.IPNet
}

type CountryResult struct {
	CountryRecord
	Network *net.IPNet
}

type ASNResult struct {
	ASNRecord
	Network *net.IPNet
}
type Config struct {
	ASN     string
	City    string
	Country string
}

type DB struct {
	cityDB    *maxminddb.Reader
	countryDB *maxminddb.Reader
	asnDB     *maxminddb.Reader
	mu        sync.RWMutex
	closed    bool
}

// Open 打开数据库
func Open(cfg Config) (*DB, error) {
	db := &DB{}
	var err error

	if cfg.City != "" {
		db.cityDB, err = maxminddb.Open(cfg.City)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Country != "" {
		db.countryDB, err = maxminddb.Open(cfg.Country)
		if err != nil {
			db.Close()
			return nil, err
		}
	}

	if cfg.ASN != "" {
		db.asnDB, err = maxminddb.Open(cfg.ASN)
		if err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}

func MustOpen(cfg Config) *DB {
	db, err := Open(cfg)
	if err != nil {
		panic(err)
	}
	return db
}

func (db *DB) LookupCity(ip net.IP) (*CityResult, error) {
	if ip == nil {
		return nil, ErrInvalidIP
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, ErrNoDB
	}

	if db.cityDB == nil {
		return nil, ErrNoDB
	}

	var rec CityRecord
	network, _, err := db.cityDB.LookupNetwork(ip, &rec)
	if err != nil {
		return nil, err
	}

	return &CityResult{
		CityRecord: rec,
		Network:    network,
	}, nil
}

func (db *DB) LookupCountry(ip net.IP) (*CountryResult, error) {
	if ip == nil {
		return nil, ErrInvalidIP
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, ErrNoDB
	}

	if db.countryDB == nil {
		return nil, ErrNoDB
	}

	var rec CountryRecord
	network, _, err := db.countryDB.LookupNetwork(ip, &rec)
	if err != nil {
		return nil, err
	}

	return &CountryResult{
		CountryRecord: rec,
		Network:       network,
	}, nil
}

func (db *DB) LookupASN(ip net.IP) (*ASNResult, error) {
	if ip == nil {
		return nil, ErrInvalidIP
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, ErrNoDB
	}

	if db.asnDB == nil {
		return nil, ErrNoDB
	}

	var rec ASNRecord
	network, _, err := db.asnDB.LookupNetwork(ip, &rec)
	if err != nil {
		return nil, err
	}

	return &ASNResult{
		ASNRecord: rec,
		Network:   network,
	}, nil
}

func (db *DB) HasCity() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.cityDB != nil
}

func (db *DB) HasCountry() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.countryDB != nil
}

func (db *DB) HasASN() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.asnDB != nil
}

func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return nil
	}
	db.closed = true

	var errs []error

	if db.cityDB != nil {
		if err := db.cityDB.Close(); err != nil {
			errs = append(errs, err)
		}
		db.cityDB = nil
	}

	if db.countryDB != nil {
		if err := db.countryDB.Close(); err != nil {
			errs = append(errs, err)
		}
		db.countryDB = nil
	}

	if db.asnDB != nil {
		if err := db.asnDB.Close(); err != nil {
			errs = append(errs, err)
		}
		db.asnDB = nil
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func PickName(names map[string]string, lang string) string {
	if names == nil {
		return ""
	}
	if v, ok := names[lang]; ok && v != "" {
		return v
	}
	if v, ok := names["en"]; ok && v != "" {
		return v
	}
	for _, v := range names {
		if v != "" {
			return v
		}
	}
	return ""
}

func IPVersion(ip net.IP) int {
	if ip.To4() != nil {
		return 4
	}
	return 6
}
