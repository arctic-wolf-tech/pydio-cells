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

// Package grpc provides the Pydio grpc service for querying indexer.
//
// Insertion in the index is not performed directly but via events broadcasted by the broker.
package grpc

import (
	"context"
	grpc2 "github.com/pydio/cells/v4/common/client/grpc"
	"path/filepath"

	"github.com/pydio/cells/v4/common/nodes/meta"

	"google.golang.org/grpc"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/broker"
	"github.com/pydio/cells/v4/common/config"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/plugins"
	"github.com/pydio/cells/v4/common/proto/sync"
	"github.com/pydio/cells/v4/common/proto/tree"
	"github.com/pydio/cells/v4/common/service"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/data/search/dao/bleve"
)

var (
	Name = common.ServiceGrpcNamespace_ + common.ServiceSearch
)

func init() {

	config.RegisterExposedConfigs(Name, ExposedConfigs)

	plugins.Register("main", func(ctx context.Context) {
		service.NewService(
			service.Name(Name),
			service.Context(ctx),
			service.Tag(common.ServiceTagData),
			service.Description("Search Engine"),
			service.Fork(true),
			/*
				service.RouterDependencies(),
				service.AutoRestart(true),
			*/
			service.WithGRPC(func(c context.Context, server *grpc.Server) error {

				cfg := servicecontext.GetConfig(c)

				indexContent := cfg.Val("indexContent").Bool()
				if indexContent {
					log.Logger(c).Info("Enabling content indexation in search engine")
				} else {
					log.Logger(c).Info("disabling content indexation in search engine")
				}

				dir, _ := config.ServiceDataDir(Name)
				bleve.IndexPath = filepath.Join(dir, "searchengine.bleve")
				bleveCfg := make(map[string]interface{})
				bleveCfg["basenameAnalyzer"] = cfg.Val("basenameAnalyzer").String()
				bleveCfg["contentAnalyzer"] = cfg.Val("contentAnalyzer").String()

				nsProvider := meta.NewNsProvider(c)

				bleveEngine, err := bleve.NewEngine(c, nsProvider, indexContent, bleveCfg)
				if err != nil {
					return err
				}

				searcher := &SearchServer{
					Engine:           bleveEngine,
					NsProvider:       nsProvider,
					TreeClient:       tree.NewNodeProviderClient(grpc2.NewClientConn(common.ServiceTree)),
					ReIndexThrottler: make(chan struct{}, 5),
				}

				tree.RegisterSearcherServer(server, searcher)
				sync.RegisterSyncEndpointServer(server, searcher)

				subscriber := searcher.Subscriber()
				if e := broker.SubscribeCancellable(c, common.TopicMetaChanges, func(message broker.Message) error {
					msg := &tree.NodeChangeEvent{}
					if ct, e := message.Unmarshal(msg); e == nil {
						return subscriber.Handle(ct, msg)
					}
					return nil
				}); e != nil {
					_ = bleveEngine.Close()
					return e
				}

				return nil
			}),
		)
	})
}