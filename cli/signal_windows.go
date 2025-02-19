// Copyright 2012-2019 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows
// +build windows

package cli

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/abmpio/abmp/pkg/log"
)

// Signal Handling
func (s *defaultCliApplication) handleSignals() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			log.Logger.Debug(fmt.Sprintf("Trapped %q signal", sig))
			s.Shutdown()
			os.Exit(0)
		}
	}()
}
