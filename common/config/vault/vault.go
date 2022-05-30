/*
 * Copyright (c) 2019-2022. Abstrium SAS <team (at) pydio.com>
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

package vault

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	vault "github.com/hashicorp/vault/api"

	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/utils/configx"
)

func init() {
	config.DefaultURLMux().Register("vault", &URLOpener{})
}

type URLOpener struct{}

func (o *URLOpener) OpenURL(ctx context.Context, u *url.URL) (config.Store, error) {
	rootToken := u.Query().Get("rootToken")
	if rootToken != "" {
		fmt.Println("Using root token from query string, this should not be used in production! You can use $VAULT_TOKEN env instead")
	}
	return New(u, rootToken)
}

func New(u *url.URL, rootToken string, opts ...configx.Option) (config.Store, error) {

	vc := vault.DefaultConfig()
	if u.Scheme == "vault" {
		vc.Address = "http://" + u.Host
	} else {
		vc.Address = "https://" + u.Host
	}

	client, err := vault.NewClient(vc)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %v", err)
	}
	if rootToken != "" {
		client.SetToken(rootToken)
	}

	return &store{
		v:      configx.New(opts...),
		cli:    client,
		prefix: strings.TrimLeft(u.Path, "/"),
	}, nil

}

type store struct {
	prefix string
	cli    *vault.Client
	v      configx.Values
}

func (s *store) Get() configx.Value {
	sec, er := s.cli.Logical().Read("secret/data/" + s.prefix)
	if er != nil {
		fmt.Println("cannot read secret " + er.Error())
	}
	i, ok := sec.Data["data"]
	if !ok {
		fmt.Println("cannot read prefix " + s.prefix)
	}
	if ms, ok := i.(map[string]interface{}); ok {
		for k, v := range ms {
			_ = s.v.Val(k).Set(v)
		}
	}
	return s.v
}

func (s *store) Set(value interface{}) error {
	if er := s.v.Set(value); er != nil {
		return er
	}
	return s.Save("internal", "autosave")
}

func (s *store) Del() error {
	_ = s.v.Del()
	_, er := s.cli.Logical().Delete("secret/data/" + s.prefix)
	return er
}

func (s *store) Val(path ...string) configx.Values {
	v := s.v.Val(path...)
	// Wrap into an autoSave configx.Values
	return &val{Values: v, store: s}
}

func (s *store) Watch(opts ...configx.WatchOption) (configx.Receiver, error) {
	return nil, fmt.Errorf("vault.watch is not implemented")
}

func (s *store) Save(s3 string, s2 string) error {
	if ms, ok := s.v.Interface().(map[string]interface{}); ok {
		_, er := s.cli.Logical().Write("secret/data/"+s.prefix, map[string]interface{}{"data": ms})
		return er
	}
	return nil
}

func (s *store) Lock() {}

func (s *store) Unlock() {}

// val wraps configx.Values to trigger save on any update
type val struct {
	configx.Values
	store *store
}

func (v *val) Set(value interface{}) error {
	if er := v.Values.Set(value); er != nil {
		return er
	}
	return v.store.Save("", "")
}

func (v *val) Del() error {
	if er := v.Values.Del(); er != nil {
		return er
	}
	return v.store.Save("", "")
}