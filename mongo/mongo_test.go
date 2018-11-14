package mongo

import (
	"fmt"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func resetMongoTimeouts(dial, retry time.Duration){
	mongoDialTimeout      = dial
	mongoDialRetryWait    = retry
}

func Test_InitSession_ShouldFailFastWhenMongoURLIsMalformed(t *testing.T) {

	defer loggertest.Reset()
	defer resetMongoTimeouts(mongoDialTimeout, mongoDialRetryWait)
	loggertest.Init(loggertest.LogLevelFatal)
	mongoDialTimeout = 100 * time.Millisecond
	mongoDialRetryWait = 100 * time.Millisecond
	c := make(chan struct{})
	url := "malformed-mongo-url?@"

	// when
	go func() {
		defer func(){c <- struct{}{}}()
		assert.Panics(t, func(){InitSession(url, 365 * 24 * 60 * 60)})
	}()

	select {
	case <-time.After(5 * time.Second):
		assert.FailNow(t, "Test has timeout while connecting to mongo")
	case <-c:
	}

	// then
	logMessages := loggertest.GetLogMessages()
	require.NotEmpty(t, logMessages)

	expectedError := fmt.Sprintf("Invalid mongo url=%s: connection option must be key=value: @", url)
	assert.Equal(t, expectedError, logMessages[0].Message)
}


func Test_InitSession_ShouldKeepTryingToConnectToMongo(t *testing.T) {
	defer loggertest.Reset()
	defer resetMongoTimeouts(mongoDialTimeout, mongoDialRetryWait)
	loggertest.Init(loggertest.LogLevelError)
	mongoDialTimeout = 100 * time.Millisecond
	mongoDialRetryWait = 100 * time.Millisecond
	c := make(chan struct{})
	url := "mongodb://localhost:27017/flyte"

	// when
	go func() {
		defer func(){c <- struct{}{}}()
		assert.NotPanics(t, func(){InitSession(url, 365 * 24 * 60 * 60)})
	}()

	select {
	case <-time.After(3 * time.Second):
	case <-c:
	}

	// then
	logMessages := loggertest.GetLogMessages()
	require.NotEmpty(t, logMessages)
	require.True(t, len(logMessages) > 1)

	expectedError := fmt.Sprintf("Unable to connect to mongo on url=%s will retry in %s: no reachable servers", url, mongoDialRetryWait)
	assert.Equal(t, expectedError, logMessages[0].Message)
	assert.Equal(t, expectedError, logMessages[1].Message)
}
