// Copyright (c) 2020 Cisco and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package up provides chain elements to 'up' interfaces (and optionally wait for them to come up)
package up

import (
	"context"
	"sync"

	"git.fd.io/govpp.git/api"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"google.golang.org/grpc"
)

type upClient struct {
	ctx        context.Context
	vppConn    Connection
	apiChannel api.Channel
	sync.Once
	initErr error
}

// NewClient provides a NetworkServiceClient chain elements that 'up's the swIfIndex
func NewClient(ctx context.Context, vppConn Connection) networkservice.NetworkServiceClient {
	return &upClient{
		ctx:     ctx,
		vppConn: vppConn,
	}
}

func (u *upClient) Request(ctx context.Context, request *networkservice.NetworkServiceRequest, opts ...grpc.CallOption) (*networkservice.Connection, error) {
	if err := u.init(ctx); err != nil {
		return nil, err
	}

	conn, err := next.Client(ctx).Request(ctx, request, opts...)
	if err != nil {
		return nil, err
	}

	if err := up(ctx, u.vppConn, u.apiChannel, true); err != nil {
		_, _ = u.Close(ctx, conn, opts...)
		return nil, err
	}
	return conn, nil
}

func (u *upClient) Close(ctx context.Context, conn *networkservice.Connection, opts ...grpc.CallOption) (*empty.Empty, error) {
	return next.Client(ctx).Close(ctx, conn, opts...)
}

func (u *upClient) init(ctx context.Context) error {
	u.Do(func() {
		u.apiChannel, u.initErr = initFunc(ctx, u.vppConn)
	})
	return u.initErr
}
