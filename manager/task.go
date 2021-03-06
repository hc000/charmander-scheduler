// The MIT License (MIT)
//
// Copyright (c) 2014 AT&T
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package manager

import (
	"encoding/json"

	"github.com/att-innovate/charmander-scheduler/mesosproto"
)

const (
	SLA_ONE_PER_NODE = "one-per-node"
	SLA_SINGLETON = "singleton"
)

type Task struct {
	ID            string    `json:"id"`
	DockerImage   string    `json:"dockerimage"`
	Mem           uint64    `json:"mem,string"`
	Cpus          float64   `json:"cpus,string"`
	Sla           string    `json:"sla"`
	NodeType      string    `json:"nodetype"`
	NodeName      string    `json:"nodename"`
	NotMetered    bool      `json:"notmetered"`
	Reshuffleable bool      `json:"reshuffleable"`
	Arguments     []string  `json:"arguments,omitempty"`
	Volumes       []*Volume `json:"volumes,omitempty"`
	Ports         []*Port   `json:"ports,omitempty"`

	InternalID    string
	SlaveID       string
	ContainerID   string
	CreatedAt     int64
	TaskInfo     *mesosproto.TaskInfo
	RequestSent   bool
	Running       bool
}

func CopyTask(source Task) Task {
	result := &Task{}

	jsonEncoded, _ := json.Marshal(&source)
	json.Unmarshal(jsonEncoded, &result)

	return *result
}

func ResetTask(task *Task) {
	task.InternalID = ""
	task.SlaveID = ""
	task.ContainerID = ""
	task.CreatedAt = 0
	task.TaskInfo = nil
	task.RequestSent = false
	task.Running = false
}
