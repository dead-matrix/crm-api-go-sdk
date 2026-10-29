package utils

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		want  time.Time
		isNil bool
	}{
		{name: "empty", in: "", isNil: true},
		{name: "garbage", in: "not-a-date", isNil: true},
		{name: "naive T is Moscow", in: "2026-10-29T15:06:49", want: time.Date(2026, 10, 29, 12, 6, 49, 0, time.UTC)},
		{name: "naive space is Moscow", in: "2026-10-29 15:06:49", want: time.Date(2026, 10, 29, 12, 6, 49, 0, time.UTC)},
		{name: "naive micro is Moscow", in: "2026-10-29T15:06:49.123456", want: time.Date(2026, 10, 29, 12, 6, 49, 123456000, time.UTC)},
		{name: "naive space micro is Moscow", in: "2026-10-29 15:06:49.5", want: time.Date(2026, 10, 29, 12, 6, 49, 500000000, time.UTC)},
		{name: "naive after midnight MSK is previous UTC day", in: "2026-10-30T01:30:00", want: time.Date(2026, 10, 29, 22, 30, 0, 0, time.UTC)},
		{name: "offset +03:00 unchanged", in: "2026-10-29T15:06:49+03:00", want: time.Date(2026, 10, 29, 12, 6, 49, 0, time.UTC)},
		{name: "Z unchanged", in: "2026-10-29T15:06:49Z", want: time.Date(2026, 10, 29, 15, 6, 49, 0, time.UTC)},
		{name: "Z with micro unchanged", in: "2026-10-29T15:06:49.123456Z", want: time.Date(2026, 10, 29, 15, 6, 49, 123456000, time.UTC)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseTime(tc.in)
			if tc.isNil {
				if got != nil {
					t.Fatalf("ParseTime(%q) = %v, want nil", tc.in, *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("ParseTime(%q) = nil, want %v", tc.in, tc.want)
			}
			if !got.Equal(tc.want) {
				t.Fatalf("ParseTime(%q) = %v, want %v", tc.in, *got, tc.want)
			}
			if got.Location() != time.UTC {
				t.Fatalf("ParseTime(%q) location = %v, want UTC", tc.in, got.Location())
			}
		})
	}
}
