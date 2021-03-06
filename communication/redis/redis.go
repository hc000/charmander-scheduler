/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package redis

import (
	"time"
	"bufio"
	"strconv"
	"net"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"

	"github.com/att-innovate/charmander-scheduler/manager"
)

func InitRedisUpdater(manager manager.Manager) {
	go updateRedis(manager)
}

func updateRedis(manager manager.Manager) {
	for {
		glog.Infoln("update redis")
		if connection := redisAvailable(manager); connection != nil {
			for _, node := range manager.GetNodes() {
				nodeInJSON, _ := json.Marshal(&node)
				key := "charmander:nodes:"+node.Hostname
				sendCommand(connection, "SET", key, fmt.Sprintf("%s", nodeInJSON))
				sendCommand(connection, "EXPIRE", key, "30") //timeout after 30s
			}
			for _, task := range manager.GetTasks() {
				if task.NotMetered { continue }
				taskInJSON, _ := json.Marshal(&task)
				key := "charmander:tasks-metered:"+task.InternalID
				sendCommand(connection, "SET", key, fmt.Sprintf("%s", taskInJSON))
				sendCommand(connection, "EXPIRE", key, "30") //timeout after 30s
			}
			for _, task := range manager.GetTasks() {
				taskInJSON, _ := json.Marshal(&task)
				key := "charmander:tasks:"+task.InternalID
				sendCommand(connection, "SET", key, fmt.Sprintf("%s", taskInJSON))
				sendCommand(connection, "EXPIRE", key, "30") //timeout after 30s
			}
			for key, value := range *getTaskIntelligence(manager) {
				parts := strings.Split(key, ":")
				manager.SetTaskIntelligence(parts[0], parts[1], value)
			}
			connection.Close()
		}

		time.Sleep(15 * time.Second)
	}
}

func getTaskIntelligence(manager manager.Manager) (*map[string]string) {
	result := map[string]string{}

	if redis := redisAvailable(manager); redis != nil {
		sendCommand(redis, "KEYS", "charmander:task-intelligence:*")
		keys := *parseResultList(redis, "")
		for _, key := range keys {
			sendCommand(redis, "GET", key)
			value := parseResult(redis, "")
			result[key[len("charmander:task-intelligence:"):]] = value
		}
		redis.Close()
	}

	return &result
}

func redisAvailable(manager manager.Manager) net.Conn {
	connection, error := net.DialTimeout("tcp", manager.GetRedisConnectionIPAndPort(), 2 * time.Second)
	if error != nil {
		return nil
	}

	return connection
}

func sendCommand(connection net.Conn, args ...string) {
	buffer := make([]byte, 0, 0)
	buffer = encodeReq(buffer, args)
	connection.Write(buffer)
}

func parseResultList(connection net.Conn, prefix string) *[]string {
	bufferedInput := bufio.NewReader(connection)
	line, _, err := bufferedInput.ReadLine()
	if err != nil {
		glog.Errorf("error parsing redis response %s\n", err)
		return &[]string {}
	}

	numberOfArgs, _ := strconv.ParseInt(string(line[1:]), 10, 64)
	args := make([]string, 0, numberOfArgs)
	for i := int64(0); i < numberOfArgs; i++ {
		line, _, _ = bufferedInput.ReadLine()
		argLen, _ := strconv.ParseInt(string(line[1:]), 10, 32)
		line, _, _ = bufferedInput.ReadLine()
		args = append(args, string(line[len(prefix):argLen]))
	}

	return &args
}

func parseResult(connection net.Conn, prefix string) string {
	bufferedInput := bufio.NewReader(connection)
	line, _, _:= bufferedInput.ReadLine()
	argLen, _ := strconv.ParseInt(string(line[1:]), 10, 32)
	line, _, _ = bufferedInput.ReadLine()

	return string(line[len(prefix):argLen])
}

func encodeReq(buf []byte, args []string) []byte {
	buf = append(buf, '*')
	buf = strconv.AppendUint(buf, uint64(len(args)), 10)
	buf = append(buf, '\r', '\n')
	for _, arg := range args {
		buf = append(buf, '$')
		buf = strconv.AppendUint(buf, uint64(len(arg)), 10)
		buf = append(buf, '\r', '\n')
		buf = append(buf, []byte(arg)...)
		buf = append(buf, '\r', '\n')
	}
	return buf
}
