package time

import gotime "time"

type Time struct {
	t *gotime.Time
}

func New() Time {
	v := gotime.Now()
	return Time{t: &v}
}

func FromTime(v gotime.Time) Time {
	return Time{t: &v}
}

func FromUnix(sec int64) Time {
	v := gotime.Unix(sec, 0)
	return Time{t: &v}
}

func FromUnixMilli(ms int64) Time {
	v := gotime.Unix(ms/1000, (ms%1000)*int64(gotime.Millisecond))
	return Time{t: &v}
}

func (ts Time) Time() gotime.Time {
	if ts.t == nil {
		return gotime.Time{}
	}
	return *ts.t
}

func (ts Time) Ptr() *gotime.Time {
	return ts.t
}

func (ts Time) Unix() int64 {
	if ts.t == nil {
		return 0
	}
	return ts.t.Unix()
}

func (ts Time) UnixMilli() int64 {
	if ts.t == nil {
		return 0
	}
	return ts.t.UnixMilli()
}

func (ts Time) Format(layout string) string {
	if ts.t == nil {
		return ""
	}
	return ts.t.Format(layout)
}

func (ts *Time) SetTime(v gotime.Time) {
	ts.t = &v
}

func (ts *Time) SetUnix(sec int64) {
	v := gotime.Unix(sec, 0)
	ts.t = &v
}

func (ts *Time) SetUnixMilli(ms int64) {
	v := gotime.Unix(ms/1000, (ms%1000)*int64(gotime.Millisecond))
	ts.t = &v
}
