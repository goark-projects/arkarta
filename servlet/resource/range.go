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
	ranges, ok, invalid := parseRanges(header, size)
	if invalid || !ok {
		return byteRange{}, false, invalid
	}
	if len(ranges) != 1 {
		return byteRange{}, false, true
	}
	return ranges[0], true, false
}

func parseRanges(header string, size int64) ([]byteRange, bool, bool) {
	if header == "" || size < 0 {
		return nil, false, false
	}
	unit, spec, ok := strings.Cut(header, "=")
	if !ok || !strings.EqualFold(strings.TrimSpace(unit), "bytes") {
		return nil, false, true
	}
	items := strings.Split(spec, ",")
	ranges := make([]byteRange, 0, len(items))
	for _, item := range items {
		target, invalid := parseByteRange(strings.TrimSpace(item), size)
		if invalid {
			return nil, false, true
		}
		ranges = append(ranges, target)
	}
	if len(ranges) == 0 {
		return nil, false, true
	}
	return ranges, true, false
}

func parseByteRange(spec string, size int64) (byteRange, bool) {
	startText, endText, ok := strings.Cut(spec, "-")
	if !ok {
		return byteRange{}, true
	}
	if startText == "" {
		suffix, err := strconv.ParseInt(endText, 10, 64)
		if err != nil || suffix <= 0 {
			return byteRange{}, true
		}
		if suffix > size {
			suffix = size
		}
		return byteRange{start: size - suffix, end: size - 1}, false
	}
	start, err := strconv.ParseInt(startText, 10, 64)
	if err != nil || start < 0 || start >= size {
		return byteRange{}, true
	}
	end := size - 1
	if endText != "" {
		end, err = strconv.ParseInt(endText, 10, 64)
		if err != nil || end < start {
			return byteRange{}, true
		}
		if end >= size {
			end = size - 1
		}
	}
	return byteRange{start: start, end: end}, false
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
