package center

import (
	"archive/zip"
	"encoding/xml"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
)

type Information struct {
	Name        string `xml:"name"`
	Version     string `xml:"version"`
	Category    string `xml:"category"`
	DisplayName string `xml:"displayName"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	AvatarUrl   string `xml:"avatarUrl"`
}

type Os struct {
	Name []string `xml:"name"`
}

type Conditions struct {
	Os         Os     `xml:"os"`
	Arch       string "xml:`arch"
	MinVersion string `xml:"min-version"`
}

type Dependencies struct {
  Dependency []string `xml:"dependency"`
}

type OptionalDependencies struct {
  OptionalDependency []string `xml:"dependency"`
}

type PluginDescriptor struct {
	Information Information `xml:"information"`
	Conditions  Conditions  `xml:"conditions"`
	Dependencies Dependencies `xml:"dependencies"`
	OptionalDependencies OptionalDependencies `xml:"optional-dependencies"`
}

func ReadDescriptor(smpFile string) (PluginDescriptor, error) {
	plugin := PluginDescriptor{}
	r, err := zip.OpenReader(smpFile)
	if err != nil {
		return plugin, errors.Wrapf(err, "failed to open smp %s", smpFile)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "META-INF/scm/plugin.xml" {
			return parseDescriptor(f)
		}
	}

	return plugin, errors.Errorf("could not find descriptor at %s", smpFile)
}

func parseDescriptor(file *zip.File) (PluginDescriptor, error) {
	plugin := PluginDescriptor{}
	rc, err := file.Open()
	if err != nil {
		return plugin, errors.Wrap(err, "failed to open zip entry")
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return plugin, errors.Wrap(err, "failed to read data from zip entry")
	}

	err = xml.Unmarshal(data, &plugin)
	if err != nil {
		log.Fatal("failed to read unmarshal plugin descriptor", err)
	}

	return plugin, nil
}
