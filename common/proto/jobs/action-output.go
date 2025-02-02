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
	"go.uber.org/zap/zapcore"

	"github.com/pydio/cells/v4/common/utils/configx"
)

type actionOutputLogArray []*ActionOutput

func (a actionOutputLogArray) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, m := range a {
		_ = encoder.AppendObject(m)
	}
	return nil
}

// JsonAsValues loads JsonBody into a configx.Values type to ease golang templating
func (m *ActionOutput) JsonAsValues() configx.Values {
	v := configx.New(configx.WithJSON())
	v.Set(m.JsonBody)
	return v
}

// JsonAsValue loads JsonBody into a configx.Value type to ease golang templating
func (m *ActionOutput) JsonAsValue() configx.Value {
	v := configx.New(configx.WithJSON())
	v.Set(m.JsonBody)
	return v.Get()
}

// MarshalLogObject implements zapcore Marshaler interface
func (m *ActionOutput) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	maxLength := 1024
	if m.Success {
		encoder.AddBool("Success", m.Success)
	}
	if m.Ignored {
		encoder.AddBool("Ignored", m.Ignored)
	}
	if m.ErrorString != "" {
		encoder.AddString("ErrorString", m.ErrorString)
	}
	if m.StringBody != "" {
		if len(m.StringBody) > maxLength {
			encoder.AddInt("StringBodyLength", len(m.StringBody))
			encoder.AddString("StringBody", m.StringBody[:maxLength]+"...")
		} else {
			encoder.AddString("StringBody", m.StringBody)
		}
	}
	if len(m.RawBody) > 0 {
		encoder.AddInt("RawBodyLength", len(m.RawBody))
	}
	if len(m.JsonBody) > 0 {
		if len(m.JsonBody) > maxLength {
			encoder.AddInt("JsonBodyLength", len(m.JsonBody))
			encoder.AddString("JsonBody", string(m.JsonBody[:maxLength])+"...")
		} else {
			bd := m.JsonAsValues().Map()
			_ = encoder.AddReflected("JsonBody", bd)
		}
	}
	return nil
}
