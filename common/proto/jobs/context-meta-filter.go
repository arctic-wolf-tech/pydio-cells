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

package jobs

import (
	"context"
	"strings"

	"github.com/ory/ladon"
	"github.com/ory/ladon/manager/memory"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/proto/idm"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	"github.com/pydio/cells/v4/common/service/context/metadata"
	"github.com/pydio/cells/v4/common/utils/permissions"
	"github.com/pydio/cells/v4/common/utils/uuid"
	"github.com/pydio/cells/v4/idm/policy/converter"
)

func (m *ContextMetaFilter) FilterID() string {
	return "ContextMetaFilter"
}

func (m *ContextMetaFilter) Filter(ctx context.Context, input ActionMessage) (ActionMessage, *ActionMessage, bool) {
	if len(m.Query.SubQueries) == 0 {
		return input, nil, true
	}
	if m.Type == ContextMetaFilterType_ContextUser {
		// Switch an IdmSelector with ContextUser as input
		a, r := m.filterContextUserQueries(ctx, input)
		return a, nil, r
	} else {
		// Apply Policy filter
		a, r := m.filterPolicyQueries(ctx, input)
		return a, nil, r
	}
}

func (m *ContextMetaFilter) filterPolicyQueries(ctx context.Context, input ActionMessage) (ActionMessage, bool) {

	policyContext := make(map[string]interface{})
	if ctxMeta, has := metadata.FromContextRead(ctx); has {
		for _, key := range []string{
			servicecontext.HttpMetaRemoteAddress,
			servicecontext.HttpMetaRequestURI,
			servicecontext.HttpMetaRequestMethod,
			servicecontext.HttpMetaUserAgent,
			servicecontext.HttpMetaContentType,
			servicecontext.HttpMetaCookiesString,
			servicecontext.HttpMetaProtocol,
			servicecontext.HttpMetaHostname,
			servicecontext.HttpMetaHost,
			servicecontext.HttpMetaPort,
			servicecontext.ServerTime,
		} {
			if val, hasKey := ctxMeta[key]; hasKey {
				policyContext[key] = val
			} else if val, hasKey := ctxMeta[strings.ToLower(key)]; hasKey {
				policyContext[key] = val
			}
		}
	}
	warden := &ladon.Ladon{Manager: memory.NewMemoryManager()}
	for _, q := range m.Query.SubQueries {
		var c ContextMetaSingleQuery
		if e := anypb.UnmarshalTo(q, &c, proto.UnmarshalOptions{}); e == nil {
			idPol := &idm.Policy{
				Id:        uuid.New(),
				Subjects:  []string{"ctx"},
				Actions:   []string{"ctx"},
				Resources: []string{"ctx"},
				Effect:    idm.PolicyEffect_allow,
				Conditions: map[string]*idm.PolicyCondition{
					c.FieldName: c.Condition,
				},
			}
			warden.Manager.Create(converter.ProtoToLadonPolicy(idPol))
		}
	}
	if err := warden.IsAllowed(&ladon.Request{
		Subject:  "ctx",
		Action:   "ctx",
		Resource: "ctx",
		Context:  policyContext,
	}); err != nil {
		log.Logger(ctx).Debug("Filter not passing : ", zap.Error(err))
		return input, false
	}
	return input, true
}

func (m *ContextMetaFilter) filterContextUserQueries(ctx context.Context, input ActionMessage) (ActionMessage, bool) {
	selector := &IdmSelector{
		Type:  IdmSelectorType_User,
		Query: m.Query,
	}
	username, _ := permissions.FindUserNameInContext(ctx)
	var user *idm.User
	if username == "" {
		log.Logger(ctx).Debug("Applying filter on ContextUser: return false as user is not found in context")
		return input, false
	}
	if username == common.PydioSystemUsername {
		user = &idm.User{Login: username}
	} else if u, err := permissions.SearchUniqueUser(ctx, username, "", &idm.UserSingleQuery{Login: username}); err == nil {
		user = u
	} else {
		log.Logger(ctx).Debug("Applying filter on ContextUser: return false as user is not found in the system")
		return input, false
	}
	// replace user
	tmpInput := input.WithUser(user)
	_, _, pass := selector.Filter(ctx, tmpInput)
	return input, pass
}
