package usecase

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2023_2_Vkladyshi/configs"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/pkg/models"
	"github.com/go-park-mail-ru/2023_2_Vkladyshi/repository/csrf"
)

type ICore interface {
	CheckCsrfToken(ctx context.Context, token string) (bool, error)
	CreateCsrfToken(ctx context.Context) (string, error)
}

type Core struct {
	csrfTokens csrf.CsrfRepo
	mutex      sync.RWMutex
	lg         *slog.Logger
}

func GetCore(cfg_sql configs.DbDsnCfg, cfg_csrf configs.DbRedisCfg, cfg_sessions configs.DbRedisCfg, lg *slog.Logger) (*Core, error) {
	csrf, err := csrf.GetCsrfRepo(cfg_csrf, lg)
	if err != nil {
		lg.Error("Csrf repository is not responding")
		return nil, err
	}

	core := Core{
		csrfTokens: *csrf,
		lg:         lg.With("module", "core"),
	}
	return &core, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (core *Core) CheckCsrfToken(ctx context.Context, token string) (bool, error) {
	core.mutex.RLock()
	found, err := core.csrfTokens.CheckActiveCsrf(ctx, token, core.lg)
	core.mutex.RUnlock()

	if err != nil {
		return false, err
	}

	return found, err
}

func (core *Core) CreateCsrfToken(ctx context.Context) (string, error) {
	sid := RandStringRunes(32)

	core.mutex.Lock()
	csrfAdded, err := core.csrfTokens.AddCsrf(
		ctx,
		models.Csrf{
			SID:       sid,
			ExpiresAt: time.Now().Add(3 * time.Hour),
		},
		core.lg,
	)
	core.mutex.Unlock()

	if !csrfAdded && err != nil {
		return "", err
	}

	if !csrfAdded {
		return "", nil
	}

	return sid, nil
}

func RandStringRunes(seed int) string {
	symbols := make([]rune, seed)
	for i := range symbols {
		symbols[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(symbols)
}
