// Copyright 2017 uSwitch
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package future

import (
	"context"
	"sync"
	"time"
)

type result struct {
	val interface{}
	err error
}

type Future struct {
	val       interface{}
	err       error
	completed bool
	sync.RWMutex
}

type FutureFn func() (interface{}, error)

func (f *Future) isComplete() bool {
	f.RLock()
	defer f.RUnlock()
	return f.completed
}

func (f *Future) Get(ctx context.Context) (interface{}, error) {
	for {
		if f.isComplete() {
			return f.val, f.err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Millisecond):
			// loop and check again
		}
	}
}

func New(f FutureFn) *Future {
	future := &Future{}
	go func() {
		val, err := f()
		future.Lock()
		future.val = val
		future.err = err
		future.completed = true
		future.Unlock()
	}()
	return future
}