package ip2

import (
	"asum/internal/task"
	"asum/internal/user"
	"asum/pkg/errorx"
	"context"
	"net"
	"sync"
)

type Network struct {
	Cidr      *string `json:"cidr,omitempty"`
	IPVersion *int    `json:"ipVersion,omitempty"`
}

type Continent struct {
	Code *string `json:"code,omitempty"`
	Name *string `json:"name,omitempty"`
}

type Country struct {
	Iso2 *string `json:"iso2,omitempty"`
	Name *string `json:"name,omitempty"`
}

type Region struct {
	Iso  *string `json:"iso,omitempty"`
	Name *string `json:"name,omitempty"`
}

type City struct {
	Name *string `json:"name,omitempty"`
}

type Postal struct {
	Code *string `json:"code,omitempty"`
}

type Location struct {
	Lat              *float64 `json:"lat,omitempty"`
	Lon              *float64 `json:"lon,omitempty"`
	AccuracyRadiusKm *int     `json:"accuracyRadiusKm,omitempty"`
}

type ASN struct {
	Number *int    `json:"number,omitempty"`
	Org    *string `json:"org,omitempty"`
}

type Traits struct {
	IsAnonymousProxy    *bool `json:"isAnonymousProxy,omitempty"`
	IsSatelliteProvider *bool `json:"isSatelliteProvider,omitempty"`
}

type GetIP struct {
	IP        string     `json:"ip"`
	Err       string     `json:"err,omitempty"`
	Network   *Network   `json:"network,omitempty"`
	Continent *Continent `json:"continent,omitempty"`
	Country   *Country   `json:"country,omitempty"`
	Region    *Region    `json:"region,omitempty"`
	City      *City      `json:"city,omitempty"`
	Postal    *Postal    `json:"postal,omitempty"`
	Location  *Location  `json:"location,omitempty"`
	Timezone  *string    `json:"timezone,omitempty"`
	Asn       *ASN       `json:"asn,omitempty"`
	Traits    *Traits    `json:"traits,omitempty"`
}

type IPItem struct {
	*GetIP
	Err   string `json:"err,omitempty"`
	Quota int64  `json:"quota,omitempty"`
}

type LookupResult struct {
	Data *GetIP
	Err  error
}

type Service interface {
	GetIP(ctx context.Context, ip, lang string) (*GetIP, error)
	BatchIP(ctx context.Context, ips []string, taskKey, lang string) (*BatchIPResp, error)
	lookupIP(ctx context.Context, ips []string, lang string) ([]*GetIP, error)
	getUserQuotaByKey(ctx context.Context, key string) int64
}

type service struct {
	repo     Repository
	userRepo user.Repository
	taskRepo task.Repository
}

func NewService(repo Repository, userRepo user.Repository, taskRepo task.Repository) Service {
	return &service{
		repo:     repo,
		userRepo: userRepo,
		taskRepo: taskRepo,
	}
}

type BatchIPResp struct {
	Quota  int64    `json:"quota"`
	Result []*GetIP `json:"result"`
}

func (s *service) GetIP(ctx context.Context, ip, lang string) (*GetIP, error) {
	result, err := s.lookupIP(ctx, []string{ip}, lang)
	return result[0], err
}

func (s *service) BatchIP(ctx context.Context, ips []string, taskKey, lang string) (*BatchIPResp, error) {
	if exist, err := s.taskRepo.ExistsByTaskKey(ctx, taskKey); err != nil || !exist {
		return nil, errorx.ErrInvalidTaskKey
	}
	quota := s.userRepo.GetQuotaByKey(ctx, taskKey)
	if len(ips) > int(quota) {
		return nil, errorx.ErrQuota
	}

	result, err := s.lookupIP(ctx, ips, lang)
	data := &BatchIPResp{Result: result, Quota: quota}
	return data, err
}

func (s *service) getUserQuotaByKey(ctx context.Context, key string) int64 {
	return s.userRepo.GetQuotaByKey(ctx, key)
}

func (s *service) lookupIP(ctx context.Context, ips []string, lang string) ([]*GetIP, error) {
	if lang == "" {
		lang = "en"
	}
	n := len(ips)
	if n == 0 {
		return nil, nil
	}

	out := make([]*GetIP, n)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 50)
	for i, ipStr := range ips {
		ipItem := net.ParseIP(ipStr)
		if ipItem == nil {
			out[i] = &GetIP{IP: ipStr, Err: errorx.ErrInvalidIP.Error()}
			continue
		}

		wg.Add(1)
		go func(idx int, ipText string, ipItem net.IP) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			data, err := s.repo.Lookup(ctx, ipItem, lang)
			if err != nil {
				out[idx] = &GetIP{IP: ipText, Err: err.Error()}
			} else {
				out[idx] = data
			}
		}(i, ipStr, ipItem)
	}

	wg.Wait()
	return out, nil
}
