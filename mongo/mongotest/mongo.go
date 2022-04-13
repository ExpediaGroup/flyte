/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mongotest

import (
	"github.com/ExpediaGroup/flyte/docker"
	"github.com/HotelsDotCom/go-logger"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2"
	"log"
	"net"
	"strconv"
	"testing"
)

type MongoT struct {
	container docker.Container
	session   *mgo.Session
	dbName    string
	host      string
	port      string
}

func NewMongoT(dbName string) *MongoT {
	port, err := getFreePort()
	if err != nil {
		logger.Fatalf("Unable to find free port for mongo: %v", err)
	}
	return &MongoT{
		host:   "localhost",
		port:   strconv.Itoa(port),
		dbName: dbName,
	}
}

func (m *MongoT) Start() {
	m.startMongoContainer(m.port + ":27017")

	session, err := mgo.Dial(m.GetUrl())
	if err != nil {
		logger.Fatalf("Unable to connect to mongo on url=%s: %v", m.GetUrl(), err)
	}

	m.session = session
}

func (m MongoT) GetSession() *mgo.Session {
	return m.session
}

func (m MongoT) GetUrl() string {
	return m.host + ":" + m.port
}

func (m *MongoT) startMongoContainer(p string) {
	d, err := docker.NewDocker()
	if err != nil {
		log.Fatal(err)
	}

	c, err := d.Run("", "docker.io/library/mongo:3.6", nil, []string{p})
	if err != nil {
		log.Fatal(err)
	}
	m.container = c
}

func (m *MongoT) Teardown() {
	if m.session != nil {
		m.session.Close()
	}

	if m.container != nil {
		if err := m.container.StopAndRemove(); err != nil {
			log.Fatal(err)
		}
	}
}

func (m MongoT) DropDatabase(t *testing.T) {
	s := m.session.Copy()
	defer s.Close()

	err := s.DB(m.dbName).DropDatabase()
	require.NoError(t, err)
}

func (m MongoT) Insert(t *testing.T, cName string, v interface{}) {
	s := m.session.Copy()
	defer s.Close()

	err := s.DB(m.dbName).C(cName).Insert(v)
	require.NoError(t, err)
}

func (m MongoT) UpsertId(t *testing.T, cName string, id, v interface{}) {
	s := m.session.Copy()
	defer s.Close()

	_, err := s.DB(m.dbName).C(cName).UpsertId(id, v)
	require.NoError(t, err)
}

func (m MongoT) Count(t *testing.T, cName string) int {
	s := m.session.Copy()
	defer s.Close()

	count, err := s.DB(m.dbName).C(cName).Count()
	require.NoError(t, err)
	return count
}

func (m MongoT) FindOne(cName string, query interface{}, v interface{}) error {
	s := m.session.Copy()
	defer s.Close()

	return s.DB(m.dbName).C(cName).Find(query).One(v)
}

func (m MongoT) FindOneT(t *testing.T, cName string, query interface{}, v interface{}) {
	err := m.FindOne(cName, query, v)
	require.NoError(t, err)
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	return l.Addr().(*net.TCPAddr).Port, l.Close()
}
