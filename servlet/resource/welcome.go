package resource

import (
	"context"
	"errors"
	"strings"
)

var defaultWelcomeFiles = []string{"index.html", "index.htm"}

// WithWelcomeFiles 设置目录请求的 welcome 文件列表；空列表表示禁用 welcome 解析。
func WithWelcomeFiles(files ...string) DefaultServletOption {
	return func(servlet *DefaultServlet) error {
		welcomeFiles, err := normalizeWelcomeFiles(files)
		if err != nil {
			return err
		}
		servlet.welcomeFiles = welcomeFiles
		return nil
	}
}

// WelcomeFiles 返回当前 welcome 文件列表副本。
func (s *DefaultServlet) WelcomeFiles() []string {
	return cloneStrings(s.welcomeFiles)
}

func (s *DefaultServlet) openWelcome(ctx context.Context, dir string) (Resource, error) {
	if len(s.welcomeFiles) == 0 {
		return Resource{}, ErrDirectory
	}
	base, _, err := cleanResourcePath(dir)
	if err != nil {
		return Resource{}, err
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	for _, welcome := range s.welcomeFiles {
		item, err := s.provider.Open(ctx, base+welcome)
		if err == nil {
			return item, nil
		}
		if !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrDirectory) {
			return Resource{}, err
		}
	}
	return Resource{}, ErrDirectory
}

func normalizeWelcomeFiles(files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	result := make([]string, 0, len(files))
	for _, file := range files {
		if file == "" || strings.HasPrefix(file, "/") {
			return nil, ErrInvalidPath
		}
		if clean, err := CleanPath(file); err != nil {
			return nil, err
		} else {
			result = append(result, strings.TrimPrefix(clean, "/"))
		}
	}
	return result, nil
}
