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

package main

import (
	"errors"
	"fmt"
	"github.com/HotelsDotCom/go-logger"
	"math"
	"os"
	"strconv"
	"strings"
)

// lookupEnv wrapper used for testing
var lookupEnv = os.LookupEnv

const (
	portEnvName                              = "FLYTE_PORT"
	tlsCertPathEnvName                       = "FLYTE_TLS_CERT_PATH"
	tlsKeyPathEnvName                        = "FLYTE_TLS_KEY_PATH"
	mgoHostEnvName                           = "FLYTE_MGO_HOST"
	authPolicyPathEnvName                    = "FLYTE_AUTH_POLICY_PATH"
	oidcIssuerURLName                        = "FLYTE_OIDC_ISSUER_URL"
	oidcIssuerClientIDName                   = "FLYTE_OIDC_ISSUER_CLIENT_ID"
	flyteTTLEnvName                          = "FLYTE_TTL_IN_SECONDS"
	shouldDeleteDeadPacksEnvName             = "FLYTE_SHOULD_DELETE_DEAD_PACKS"
	deleteDeadPacksTimeEnvName	             = "FLYTE_DELETE_DEAD_PACKS_AT_HH_COLON_MM"
	packGracePeriodUntilDeadInSecondsEnvName = "FLYTE_PACK_GRACE_PERIOD_UNTIL_MARKED_DEAD_IN_SECONDS"
	defaultDeleteDeadPacksTime				 = "23:00"
	oneWeekInSeconds                         = 604800
	oneYearInSeconds                         = 31557600
)

type Config struct {
	MongoHost          				  	string
	Port               				  	string
	TLSCertPath        				  	string
	TLSKeyPath         				  	string
	AuthPolicyPath     				  	string
	OidcIssuerURL      				  	string
	OidcIssuerClientID 				  	string
	FlyteTTL           				  	int
	ShouldDeleteDeadPacks				bool
	DeleteDeadPacksTime					string
	PackGracePeriodUntilDeadInSeconds	int
}

func NewConfig() Config {
	c := Config{}
	c.MongoHost = getEnvVarWithDefault(mgoHostEnvName, "localhost:27017")
	c.TLSCertPath = getPathVar(tlsCertPathEnvName)
	c.TLSKeyPath = getPathVar(tlsKeyPathEnvName)
	c.Port = c.getPort()
	c.AuthPolicyPath = getEnvVar(authPolicyPathEnvName)
	c.OidcIssuerURL = getEnvVar(oidcIssuerURLName)
	c.OidcIssuerClientID = getEnvVar(oidcIssuerClientIDName)
	c.FlyteTTL = getIntEnvVarWithDefault(flyteTTLEnvName, oneYearInSeconds)
	c.ShouldDeleteDeadPacks = getBoolEnvVarWithDefault(shouldDeleteDeadPacksEnvName, false)
	c.DeleteDeadPacksTime = getDeleteDeadPacksTimeEnvVarWithDefault(deleteDeadPacksTimeEnvName, defaultDeleteDeadPacksTime)
	c.PackGracePeriodUntilDeadInSeconds = getIntEnvVarWithDefault(packGracePeriodUntilDeadInSecondsEnvName, oneWeekInSeconds)
	return c
}

func getEnvVarWithDefault(name, defaultVal string) string {
	val, isSet := lookupEnv(name)
	if !isSet {
		logger.Infof(fmt.Sprintf("%s env not set, using default", name))
		val = defaultVal
	}
	logger.Infof("Using %s=%s", name, val)
	return val
}

func getIntEnvVarWithDefault(name string, defaultVal int) int {
	val, isSet := lookupEnv(name)
	if !isSet {
		logger.Infof(fmt.Sprintf("%s env not set, using default", name))
		return defaultVal
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		logger.Errorf(fmt.Sprintf("Error converting %s to int, using default. Value of %s: %v", name, name, val))
		return defaultVal
	}

	logger.Infof("Using %s=%v", name, intVal)
	return intVal
}

func getBoolEnvVarWithDefault(name string, defaultVal bool) bool {
	val, isSet := lookupEnv(name)
	if !isSet {
		logger.Infof(fmt.Sprintf("%s env not set, using default: %v", name, defaultVal))
		return defaultVal
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		logger.Errorf(fmt.Sprintf("Error converting %s to bool, using default: %v. Value of %s: %v", name, defaultVal, name, val))
		return defaultVal
	}

	logger.Infof("Using %s=%v", name, boolVal)
	return boolVal
}

func getPathVar(envName string) string {
	path := getEnvVar(envName)
	if path != "" && !fileExists(path) {
		logger.Fatalf("cannot find file defined by: %v=%v", envName, path)
	}
	return path
}

func getEnvVar(name string) string {
	val, isSet := lookupEnv(name)
	if isSet {
		logger.Infof("Using %s=%s", name, val)
	}
	return val
}

var fileExists = func(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c Config) getPort() string {
	port := getEnvVarWithDefault(portEnvName, c.getDefaultPort())
	if portNumber, err := strconv.Atoi(port); err != nil || !isValidPortNumber(portNumber) {
		logger.Fatalf("invalid port: %v=%v", portEnvName, port)
	}
	return port
}

func (c Config) getDefaultPort() string {
	if c.requireTLS() {
		return "8443"
	}
	return "8080"
}

func isValidPortNumber(port int) bool {
	if port < 0 || port > math.MaxUint16 {
		return false
	}
	return true
}

func (c Config) requireTLS() bool {
	return c.TLSCertPath != "" && c.TLSKeyPath != ""
}

func (c Config) requireAuth() bool {
	return c.AuthPolicyPath != "" && c.OidcIssuerURL != "" && c.OidcIssuerClientID != ""
}

func getDeleteDeadPacksTimeEnvVarWithDefault(name, defaultVal string) string {
	val, isSet := lookupEnv(name)
	if !isSet {
		logger.Infof(fmt.Sprintf("%s env not set, using default %v", name, defaultVal))
		return defaultVal
	}
	if err := validateTime(val); err != nil {
		logger.Errorf(fmt.Sprintf("%s env is invalid, using default %v, error: %v", name, defaultVal, err))
		return defaultVal
	}

	logger.Infof("Using %s=%s", name, val)
	return val
}

func validateTime(t string) error {
	var err error
	var hour int
	var min int
	ts := strings.Split(t, ":")
	if len(ts) != 2 {
		return errors.New(fmt.Sprintf("time format error. time is not in 'HH:MM' format. invalid value: %s.", t))
	}
	if hour, err = strconv.Atoi(ts[0]); err != nil {
		return errors.New(fmt.Sprintf("time format error. hour is invalid. invalid value: %s. err: %v.", t, err))
	}
	if min, err = strconv.Atoi(ts[1]); err != nil {
		return errors.New(fmt.Sprintf("time format error. minute is invalid. invalid value: %s. err: %v.", t, err))
	}
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		return errors.New(fmt.Sprintf("time format error. hours should be 0-23, minute should be 0-59. invalid value: %s.", t))
	}
	return nil
}
