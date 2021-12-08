/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package auth

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/logrusx"
	"github.com/pydio/cells/v4/common/log"
	"go.uber.org/zap"

	"github.com/ory/fosite"
	"github.com/ory/hydra/client"
	"github.com/ory/hydra/driver"
	"github.com/ory/x/sqlcon"
	"github.com/pkg/errors"
	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/proto/install"
	"github.com/pydio/cells/v4/common/sql"
	"github.com/sirupsen/logrus"
)

var (
	reg  driver.Registry
	once = &sync.Once{}

	syncLock = &sync.Mutex{}

	onRegistryInits []func()
)

func InitRegistry(dao sql.DAO) {
	once.Do(func() {

		// db := sqlx.NewDb(dao.DB(), dao.Driver())
		testL := logrus.New()
		testL.SetOutput(io.Discard)
		lx := logrusx.New("test", "1", logrusx.UseLogger(testL))

		var e error
		cfg := defaultConf.GetProvider()
		cfg.Set("dsn", "mysql://root@tcp(localhost:3306)/cells?parseTime=true")
		reg, e = driver.NewRegistryFromDSN(context.Background(), cfg, lx)
		if e != nil {
			fmt.Printf("Cannot init registryFromDSN", e)
			os.Exit(1)
		}
		/*
			reg = driver.New(
				context.Background(),
				driver.WithOptions(
					configx.SkipValidation(),
					configx.WithLogger(lx),
					// TODO V4
					configx.WithValue("dsn", "mysql://root@tcp(localhost:3306)/cells?parseTime=true"),
				),
				driver.DisableValidation(),
				driver.DisablePreloading())
		*/
		p := reg.WithConfig(defaultConf.GetProvider()).Persister()
		conn := p.Connection(context.Background())

		if err := conn.Open(); err != nil {
			fmt.Printf("Could not open the database connection:\n%+v\n", err)
			os.Exit(1)
			return
		}

		// convert migration tables
		if err := p.PrepareMigration(context.Background()); err != nil {
			fmt.Printf("Could not convert the migration table:\n%+v\n", err)
			os.Exit(1)
			return
		}

		// print migration status
		//fmt.Println("The following migration is planned:")
		//fmt.Println("")

		_, err := p.MigrationStatus(context.Background())
		if err != nil {
			fmt.Printf("Could not get the migration status:\n%+v\n", errorsx.WithStack(err))
			os.Exit(1)
			return
		}
		//_ = status.Write(os.Stdout)

		// apply migrations
		fmt.Println("Applying migrations for oauth if required")
		if err := p.MigrateUp(context.Background()); err != nil {
			fmt.Printf("Could not apply migrations:\n%+v\n", errorsx.WithStack(err))
		}
		fmt.Println("Finished")

		RegisterOryProvider(reg.WithConfig(defaultConf.GetProvider()).OAuth2Provider())
	})

	if err := syncClients(context.Background(), reg.ClientManager(), defaultConf.Clients()); err != nil {
		log.Warn("Failed to sync clients", zap.Error(err))
		return
	}

	for _, onRegistryInit := range onRegistryInits {
		onRegistryInit()
	}
}

func OnRegistryInit(f func()) {
	onRegistryInits = append(onRegistryInits, f)
}

func GetRegistry() driver.Registry {
	return reg
}

func DuplicateRegistryForConf(c ConfigurationProvider) driver.Registry {
	l := logrus.New()
	l.SetLevel(logrus.PanicLevel)
	return driver.NewRegistrySQL() //TODO V4 .WithConfig(c).WithLogger(l)
}

func GetRegistrySQL() *driver.RegistrySQL {
	return reg.(*driver.RegistrySQL)
}

func syncClients(ctx context.Context, s client.Storage, c common.Scanner) error {
	var clients []*client.Client

	if c == nil {
		return nil
	}

	syncLock.Lock()
	defer syncLock.Unlock()

	if err := c.Scan(&clients); err != nil {
		return err
	}

	n, err := s.CountClients(ctx)
	if err != nil {
		return err
	}

	var old []client.Client
	if n > 0 {
		if o, err := s.GetClients(ctx, client.Filter{Offset: 0, Limit: n}); err != nil {
			return err
		} else {
			old = o
		}
	}
	sites, _ := config.LoadSites()

	for _, cli := range clients {
		_, err := s.GetClient(ctx, cli.GetID())

		var redirectURIs []string
		for _, r := range cli.RedirectURIs {
			tt := rangeFromStr(r)
			for _, t := range tt {
				vv := varsFromStr(t, sites)
				redirectURIs = append(redirectURIs, vv...)
			}
		}

		cli.RedirectURIs = redirectURIs

		if errors.Cause(err) == sqlcon.ErrNoRows {
			// Let's create it
			if err := s.CreateClient(ctx, cli); err != nil {
				return err
			}
		} else {
			if err := s.UpdateClient(ctx, cli); err != nil {
				return err
			}
		}

		var cleanOld []client.Client
		for _, o := range old {
			if o.GetID() == cli.GetID() {
				continue
			}
			cleanOld = append(cleanOld, o)
		}
		old = cleanOld
	}

	for _, cli := range old {
		if err := s.DeleteClient(ctx, cli.GetID()); err != nil {
			return err
		}
	}

	return nil
}

func rangeFromStr(s string) []string {

	var res []string
	re := regexp.MustCompile(`\[([0-9]+)-([0-9]+)\]`)

	r := re.FindStringSubmatch(s)

	if len(r) < 3 {
		return []string{s}
	}

	min, err := strconv.Atoi(r[1])
	if err != nil {
		return []string{s}
	}

	max, err := strconv.Atoi(r[2])
	if err != nil {
		return []string{s}
	}

	if min > max {
		return []string{s}
	}

	for {
		if min > max {
			break
		}

		res = append(res, re.ReplaceAllString(s, strconv.Itoa(min)))

		min = min + 1
	}
	return res
}

func varsFromStr(s string, sites []*install.ProxyConfig) []string {
	var res []string
	defaultBind := ""
	if len(sites) > 0 {
		defaultBind = config.GetDefaultSiteURL(sites...)
	}
	if strings.Contains(s, "#default_bind#") {

		res = append(res, strings.ReplaceAll(s, "#default_bind#", defaultBind))

	} else if strings.Contains(s, "#binds...#") {

		for _, si := range sites {
			for _, u := range si.GetExternalUrls() {
				res = append(res, strings.ReplaceAll(s, "#binds...#", u.String()))
			}
		}

	} else if strings.Contains(s, "#insecure_binds...") {

		for _, si := range sites {
			for _, u := range si.GetExternalUrls() {
				if !fosite.IsRedirectURISecure(u) {
					res = append(res, strings.ReplaceAll(s, "#insecure_binds...#", u.String()))
				}
			}
		}

	} else {

		res = append(res, s)

	}
	return res
}