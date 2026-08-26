package resource

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type byteRange struct {
	start int64
	end   int64
}

func parseRange(header string, size int64) (byteRange, bool, bool) {
	if header == "" || size < 0 {
		return byteRange{}, false, false
	}
	unit, spec, ok := strings.Cut(header, "=")
	if !ok || !strings.EqualFold(strings.TrimSpace(unit), "bytes") || strings.Contains(spec, ",") {
		return byteRange{}, false, true
	}
	startText, endText, ok := strings.Cut(strings.TrimSpace(spec), "-")
	if !ok {
		return byteRange{}, false, true
	}
	if startText == "" {
		suffix, err := strconv.ParseInt(endText, 10, 64)
		if err != nil || suffix <= 0 {
			return byteRange{}, false, true
		}
		if suffix > size {
			suffix = size
		}
		return byteRange{start: size - suffix, end: size - 1}, true, false
	}
	start, err := strconv.ParseInt(startText, 10, 64)
	if err != nil || start < 0 || start >= size {
		return byteRange{}, false, true
	}
	end := size - 1
	if endText != "" {
		end, err = strconv.ParseInt(endText, 10, 64)
		if err != nil || end < start {
			return byteRange{}, false, true
		}
		if end >= size {
			end = size - 1
		}
	}
	return byteRange{start: start, end: end}, true, false
}

func (r byteRange) length() int64 {
	return r.end - r.start + 1
}

func (r byteRange) contentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.end, size)
}

func unsatisfiedRange(size int64) string {
	return fmt.Sprintf("bytes */%d", size)
}

func rangeNotSatisfiable() int {
	return http.StatusRequestedRangeNotSatisfiable
}
