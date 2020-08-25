package pack

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPack_SetStatus(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name     string
		lastSeen time.Time
		want     string
	}{
		{lastSeen: now, want: "live"},
		{lastSeen: now.Add(-(10*60 - 1) * time.Second), want: "live"},
		{lastSeen: now.Add(-10 * time.Minute), want: "warning"},
		{lastSeen: now.Add(-(24*60 - 1) * time.Minute), want: "warning"},
		{lastSeen: now.Add(-24 * time.Hour), want: "critical"},
		{lastSeen: now.Add(-(24 * 356) * time.Hour), want: "critical"},
	}

	for _, tt := range cases {
		t.Run(tt.want+"_duration:"+now.Sub(tt.lastSeen).String(), func(t *testing.T) {
			p := packResponse{Pack: Pack{LastSeen: tt.lastSeen}}
			p.setStatus()
			assert.Equal(t, tt.want, p.Status)
		})
	}
}
