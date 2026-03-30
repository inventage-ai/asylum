package docker

import "testing"

func TestCountExecSessions(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   int
	}{
		{
			"empty",
			"",
			0,
		},
		{
			"only PID 1 and check command",
			"    1     0\n   42     0\n",
			1, // 42 is the check command
		},
		{
			"one other session plus check",
			"    1     0\n   38     0\n   42     0\n",
			2,
		},
		{
			"background tasks ignored",
			"    1     0\n    7     1\n  314     1\n   42     0\n",
			1, // only 42 (check), 7 and 314 have PPID=1
		},
		{
			"two sessions plus background tasks",
			"    1     0\n    7     1\n   38     0\n  115     0\n  314     1\n   42     0\n",
			3, // 38, 115, 42
		},
		{
			"child processes of sessions ignored",
			"    1     0\n   38     0\n  114    38\n   42     0\n",
			2, // 38 and 42; 114 has PPID=38
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countExecSessions(tt.input)
			if got != tt.want {
				t.Errorf("countExecSessions() = %d, want %d", got, tt.want)
			}
		})
	}
}
