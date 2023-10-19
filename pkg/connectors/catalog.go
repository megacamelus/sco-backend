package connectors

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"

	"github.com/sco1237896/sco-backend/pkg/logger"
	"github.com/sco1237896/sco-backend/pkg/utils/files"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	DefaultCatalogLocation = "/etc/connectors"
)

type CatalogOptions struct {
	Dirs []string
}

func DefaultCatalogOptions() CatalogOptions {
	return CatalogOptions{
		Dirs: []string{DefaultCatalogLocation},
	}
}

// TODO: make more efficient
// TODO: ideally we should use values instead of pointers

type Catalog struct {
	Connectors []*Connector
	ByName     map[string]*Connector
}

func NewCatalog(opts CatalogOptions) (*Catalog, error) {
	c := Catalog{}
	c.Connectors = make([]*Connector, 0)
	c.ByName = make(map[string]*Connector)

	for i := range opts.Dirs {
		if _, err := os.Stat(opts.Dirs[i]); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, err
		}

		err := files.Walk(opts.Dirs[i], func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// skip over hidden files and dirs..
			if info.IsDir() || strings.HasPrefix(path, ".") {
				return nil
			}

			logger.L.Debug("loading connectors", slog.String("file", path))

			// Read the file
			buf, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading catalog file %s: %w", path, err)
			}

			u := unstructured.Unstructured{}
			if err := decode(buf, &u); err != nil {
				return fmt.Errorf("error unmarshaling catalog file %s: %w", path, err)
			}

			// skip non definition files
			if u.GetKind() != "ConnectorDefinition" {
				return nil
			}

			connector := Connector{}
			if err := decode(buf, &connector); err != nil {
				return fmt.Errorf("error unmarshaling connector definition file %s: %w", path, err)
			}

			logger.L.Debug("loaded connector", slog.String("file", path), slog.String("name", connector.Name))

			c.Connectors = append(c.Connectors, &connector)
			c.ByName[connector.Name] = &connector

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	logger.L.Info("catalog loaded", slog.Int("connectors", len(c.Connectors)))

	return &c, nil
}

func decode(data []byte, out interface{}) error {
	reader := bytes.NewReader(data)
	decoder := yaml.NewYAMLToJSONDecoder(reader)

	return decoder.Decode(out)
}
