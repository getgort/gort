/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package adapter

import (
	"context"

	"github.com/getgort/gort/data/rest"
)

var (
	contextChatChannel = Key{"gort.context.chat.channel"}
	contextChatUser    = Key{"gort.context.chat.user"}
	contextGortGroup   = Key{"gort.context.gort.group"}
	contextGortUser    = Key{"gort.context.gort.user"}
)

type Key struct {
	key string
}

func (k *Key) String() string {
	return k.key
}

func WithChatChannel(ctx context.Context, channel *ChannelInfo) context.Context {
	return context.WithValue(ctx, contextChatChannel, channel)
}

func WithChatUser(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, contextChatUser, user)
}

// func WithGortGroup(ctx context.Context, group *rest.Group) context.Context {
// 	return context.WithValue(ctx, contextChatUser, group)
// }

func WithGortUser(ctx context.Context, user rest.User) context.Context {
	return context.WithValue(ctx, contextGortUser, user)
}

func GetChatChannel(ctx context.Context) (*ChannelInfo, bool) {
	if i := ctx.Value(contextChatChannel); i != nil {
		return i.(*ChannelInfo), true
	}
	return nil, false
}

func GetChatUser(ctx context.Context) (*UserInfo, bool) {
	if i := ctx.Value(contextChatUser); i != nil {
		return i.(*UserInfo), true
	}
	return nil, false
}

// func GetGortGroup(ctx context.Context) *rest.Group {
// 	if i := ctx.Value(contextGortGroup); i != nil {
// 		return i.(*rest.Group)
// 	}
// 	return nil
// }

func GetGortUser(ctx context.Context) (rest.User, bool) {
	if i := ctx.Value(contextGortUser); i != nil {
		return i.(rest.User), true
	}
	return rest.User{}, false
}
