// Copyright 2012 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"fmt"
	"github.com/globocom/commandmocker"
	"github.com/globocom/tsuru/api/bind"
	"github.com/globocom/tsuru/provision"
	"github.com/globocom/tsuru/repository"
	. "launchpad.net/gocheck"
	"strings"
)

func (s *S) TestCommand(c *C) {
	tmpdir, err := commandmocker.Add("juju", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	u := Unit{
		Type:    "django",
		Name:    "i-0800",
		Machine: 1,
		State:   provision.StatusStarted,
		app:     &App{},
	}
	var buf bytes.Buffer
	err = u.Command(&buf, &buf, "uname")
	c.Assert(err, IsNil)
	c.Assert(buf.String(), Matches, `.* \d uname`)
}

func (s *S) TestCommandShouldAcceptMultipleParams(c *C) {
	dir, err := commandmocker.Add("juju", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(dir)
	u := Unit{
		Type:    "django",
		Name:    "myUnit",
		Machine: 1,
		State:   provision.StatusStarted,
		app:     &App{},
	}
	var buf bytes.Buffer
	err = u.Command(&buf, &buf, "uname", "-a")
	c.Assert(buf.String(), Matches, `.* \d uname -a`)
}

func (s *S) TestCommandWithCustomStdout(c *C) {
	dir, err := commandmocker.Add("juju", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(dir)
	u := Unit{
		Type:    "django",
		Name:    "myUnit",
		Machine: 1,
		State:   provision.StatusStarted,
		app:     &App{},
	}
	var b bytes.Buffer
	u.Command(&b, nil, "uname", "-a")
	c.Assert(b.String(), Matches, `.* \d uname -a`)
}

func (s *S) TestCommandWithCustomStderr(c *C) {
	dir, err := commandmocker.Error("juju", "$*", 42)
	c.Assert(err, IsNil)
	defer commandmocker.Remove(dir)
	u := Unit{
		Type:    "django",
		Name:    "myUnit",
		Machine: 1,
		State:   provision.StatusStarted,
		app:     &App{},
	}
	var b bytes.Buffer
	err = u.Command(nil, &b, "uname", "-a")
	c.Assert(err, NotNil)
	c.Assert(b.String(), Matches, `.* \d uname -a`)
}

func (s *S) TestCommandReturnErrorIfTheUnitIsNotStarted(c *C) {
	u := Unit{
		Type:    "django",
		Name:    "myUnit",
		Machine: 1,
		State:   provision.StatusInstalling,
		app:     &App{},
	}
	err := u.Command(nil, nil, "uname", "-a")
	c.Assert(err, NotNil)
	expected := fmt.Sprintf("Unit must be started to run commands, but it is %q.", u.State)
	c.Assert(err.Error(), Equals, expected)
}

func (s *S) TestWriteEnvVars(c *C) {
	tmpdir, err := commandmocker.Add("juju", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	app := App{
		Name: "intheend",
		Env: map[string]bind.EnvVar{
			"https_proxy": {
				Name:   "https_proxy",
				Value:  "https://secureproxy.com:3128/",
				Public: true,
			},
		},
	}
	unit := Unit{
		Type:    "django",
		Name:    "myunit",
		Machine: 50,
		State:   provision.StatusStarted,
		app:     &app,
	}
	err = unit.writeEnvVars()
	c.Assert(err, IsNil)
	c.Assert(commandmocker.Ran(tmpdir), Equals, true)
	output := strings.Replace(commandmocker.Output(tmpdir), "\n", " ", -1)
	outputRegexp := `^.*50 cat > /home/application/apprc <<END # generated by tsuru .*`
	outputRegexp += ` export https_proxy="https://secureproxy.com:3128/" END $`
	c.Assert(output, Matches, outputRegexp)
}

func (s *S) TestWriteEnvVarsErrorWithoutOutput(c *C) {
	tmpdir, err := commandmocker.Add("juju", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	app := App{
		Name: "intheend",
		Env: map[string]bind.EnvVar{
			"https_proxy": {
				Name:   "https_proxy",
				Value:  "https://secureproxy.com:3128/",
				Public: true,
			},
		},
	}
	unit := Unit{
		Type:    "django",
		Name:    "myunit",
		Machine: 50,
		app:     &app,
	}
	err = unit.writeEnvVars()
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `^Failed to write env vars: Unit must be started.*\.$`)
	c.Assert(commandmocker.Ran(tmpdir), Equals, false)
}

func (s *S) TestWriteEnvVarsErrorWithOutput(c *C) {
	tmpdir, err := commandmocker.Error("juju", "juju failed", 1)
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	app := App{
		Name: "intheend",
		Env: map[string]bind.EnvVar{
			"https_proxy": {
				Name:   "https_proxy",
				Value:  "https://secureproxy.com:3128/",
				Public: true,
			},
		},
	}
	unit := Unit{
		Type:    "django",
		Name:    "myunit",
		Machine: 50,
		State:   provision.StatusStarted,
		app:     &app,
	}
	err = unit.writeEnvVars()
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `^Failed to write env vars \(exit status 1\): juju failed\.$`)
	c.Assert(commandmocker.Ran(tmpdir), Equals, true)
}

func (s *S) TestUnitGetName(c *C) {
	u := Unit{Name: "abcdef", app: &App{Name: "2112"}}
	c.Assert(u.GetName(), Equals, "abcdef")
}

func (s *S) TestUnitGetMachine(c *C) {
	u := Unit{Machine: 10}
	c.Assert(u.GetMachine(), Equals, u.Machine)
}

func (s *S) TestUnitShouldBeARepositoryUnit(c *C) {
	var _ repository.Unit = &Unit{}
}

func (s *S) TestUnitShouldBeABinderUnit(c *C) {
	var _ bind.Unit = &Unit{}
}
