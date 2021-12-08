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

package broker

import (
	"context"

	"github.com/micro/micro/v3/service/broker"
	"google.golang.org/protobuf/proto"

	"github.com/pydio/cells/v4/common/service/context/metadata"
)

type UnSubscriber func() error

type SubscriberHandler func(Message) error

type Message interface {
	Unmarshal(target proto.Message) (context.Context, error)
}

type SubMessage struct {
	*broker.Message
}

func (m *SubMessage) Unmarshal(target proto.Message) (context.Context, error) {
	if e := proto.Unmarshal(m.Body, target); e != nil {
		return nil, e
	}
	ctx := context.Background()
	if m.Header != nil {
		ctx = metadata.NewContext(ctx, m.Header)
	}
	return ctx, nil
}