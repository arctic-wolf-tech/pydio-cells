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

package chat

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/pydio/cells/v4/common/dao/mongodb"
	"github.com/pydio/cells/v4/common/proto/chat"
	"github.com/pydio/cells/v4/common/utils/configx"
	"github.com/pydio/cells/v4/common/utils/uuid"
)

func TestMongoDAO_Init(t *testing.T) {

	mDsn := os.Getenv("CELLS_TEST_MONGODB_DSN")
	if mDsn == "" {
		t.Log("Skipping Mongodb Test, no URI set")
		return
	}

	h, _ := mongodb.NewDAO("mongodb", mDsn, "chat-test")
	m := NewDAO(h).(DAO)
	conf := configx.New()
	er := m.Init(conf)
	if er != nil {
		t.Error(er)
	}

	Convey("Test connection init", t, func() {
		nodeUuid := uuid.New()
		roomUuid := uuid.New()

		_, e := m.PutRoom(&chat.ChatRoom{
			Type:           chat.RoomType_NODE,
			Uuid:           roomUuid,
			RoomLabel:      "My Room",
			RoomTypeObject: nodeUuid,
			LastUpdated:    int32(time.Now().Unix()),
			Users:          []string{"admin", "user"},
		})

		So(e, ShouldBeNil)

		cc, e := m.ListRooms(&chat.ListRoomsRequest{ByType: chat.RoomType_NODE})
		So(e, ShouldBeNil)
		So(len(cc), ShouldBeGreaterThanOrEqualTo, 1)
		So(cc[len(cc)-1].Uuid, ShouldEqual, roomUuid)

		cc, e = m.ListRooms(&chat.ListRoomsRequest{ByType: chat.RoomType_NODE, TypeObject: nodeUuid})
		So(e, ShouldBeNil)
		So(cc, ShouldHaveLength, 1)
		So(cc[0].Uuid, ShouldEqual, roomUuid)

		res, e := m.RoomByUuid(chat.RoomType_NODE, roomUuid)
		So(e, ShouldBeNil)
		So(res.Uuid, ShouldEqual, roomUuid)

		msg, e := m.PostMessage(&chat.ChatMessage{Message: "Hello world", RoomUuid: roomUuid, Timestamp: time.Now().UnixNano()})
		So(e, ShouldBeNil)
		So(msg.Uuid, ShouldNotBeEmpty)

		msg2, e := m.PostMessage(&chat.ChatMessage{Message: "Hello world 2", RoomUuid: roomUuid, Timestamp: time.Now().UnixNano()})
		So(e, ShouldBeNil)
		So(msg2.Uuid, ShouldNotBeEmpty)

		// List messages sorted
		mm, e := m.ListMessages(&chat.ListMessagesRequest{RoomUuid: roomUuid})
		So(e, ShouldBeNil)
		So(mm, ShouldHaveLength, 2)
		So(mm[0].Uuid, ShouldEqual, msg.Uuid)
		So(mm[0].Message, ShouldEqual, "Hello world")

		e = m.DeleteMessage(&chat.ChatMessage{Uuid: msg.Uuid})
		So(e, ShouldBeNil)

		mm2, e := m.ListMessages(&chat.ListMessagesRequest{RoomUuid: roomUuid})
		So(e, ShouldBeNil)
		So(mm2, ShouldHaveLength, 1)
		So(mm2[0].Uuid, ShouldEqual, msg2.Uuid)
		So(mm2[0].Message, ShouldEqual, "Hello world 2")

		ok, er := m.DeleteRoom(&chat.ChatRoom{Uuid: roomUuid})
		So(er, ShouldBeNil)
		So(ok, ShouldBeTrue)

	})

	Convey("Clean DB", t, func() {
		for i := 0; i < 4; i++ {
			cc, e := m.ListRooms(&chat.ListRoomsRequest{ByType: chat.RoomType(i)})
			So(e, ShouldBeNil)
			for _, c := range cc {
				_, e = m.DeleteRoom(c)
				So(e, ShouldBeNil)
			}
		}
	})
}