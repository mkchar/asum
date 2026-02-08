package ip2

import (
	"context"
	"net"

	"asum/pkg/maxmind"
)

type Repository interface {
	Lookup(ctx context.Context, ip net.IP, lang string) (*GetIP, error)
	Close() error
}

type repository struct {
	*maxmind.DB
}

func NewRepository(db *maxmind.DB) Repository {
	return &repository{DB: db}
}

func (r *repository) Lookup(ctx context.Context, ip net.IP, lang string) (*GetIP, error) {
	_ = ctx

	out := &GetIP{
		IP: ip.String(),
	}
	ipVer := int(maxmind.IPVersion(ip))
	out.Network = &Network{
		IPVersion: &ipVer,
	}

	if r.HasCity() {
		city, err := r.LookupCity(ip)
		if err == nil && city != nil {
			r.fromCity(out, city, lang)
		}
	}

	if r.HasCountry() && r.needCountryFallback(out) {
		country, err := r.LookupCountry(ip)
		if err == nil && country != nil {
			r.fromCountry(out, country, lang)
		}
	}

	if r.HasASN() {
		asn, err := r.LookupASN(ip)
		if err == nil && asn != nil {
			out.Asn = &ASN{
				Number: &asn.Number,
				Org:    &asn.Org,
			}
		}
	}

	return out, nil
}

func (r *repository) fromCity(out *GetIP, city *maxmind.CityResult, lang string) {
	if city.Network != nil {
		cidr := city.Network.String()
		out.Network.Cidr = &cidr
	}

	if city.Continent.Code != "" {
		continentName := maxmind.PickName(city.Continent.Names, lang)
		out.Continent = &Continent{
			Code: &city.Continent.Code,
			Name: &continentName,
		}
	}

	if city.Country.ISOCode != "" {
		countryName := maxmind.PickName(city.Country.Names, lang)
		out.Country = &Country{
			Iso2: &city.Country.ISOCode,
			Name: &countryName,
		}
	}
	if len(city.Subdivisions) > 0 {
		subdivisionsName := maxmind.PickName(city.Subdivisions[0].Names, lang)
		out.Region = &Region{
			Iso:  &city.Subdivisions[0].ISOCode,
			Name: &subdivisionsName,
		}
	}

	cityName := maxmind.PickName(city.City.Names, lang)
	if cityName != "" {
		out.City = &City{
			Name: &cityName,
		}
	}

	if city.Postal.Code != "" {
		out.Postal = &Postal{
			Code: &city.Postal.Code,
		}
	}
	if city.Location.Latitude != 0 || city.Location.Longitude != 0 {
		out.Location = &Location{
			Lat:              &city.Location.Latitude,
			Lon:              &city.Location.Longitude,
			AccuracyRadiusKm: &city.Location.AccuracyRadius,
		}
	}

	if city.Location.TimeZone != "" {
		out.Timezone = &city.Location.TimeZone
	}

	out.Traits = &Traits{
		IsAnonymousProxy:    &city.Traits.IsAnonymousProxy,
		IsSatelliteProvider: &city.Traits.IsSatelliteProvider,
	}
}

func (r *repository) fromCountry(out *GetIP, country *maxmind.CountryResult, lang string) {
	if out.Continent == nil && country.Continent.Code != "" {
		continentName := maxmind.PickName(country.Continent.Names, lang)
		out.Continent = &Continent{
			Code: &country.Continent.Code,
			Name: &continentName,
		}
	}

	if out.Country == nil && country.Country.ISOCode != "" {
		countryName := maxmind.PickName(country.Country.Names, lang)
		out.Country = &Country{
			Iso2: &country.Country.ISOCode,
			Name: &countryName,
		}
	}
}

func (r *repository) needCountryFallback(out *GetIP) bool {
	return out.Country == nil || out.Continent == nil
}

func (r *repository) Close() error {
	return r.DB.Close()
}
