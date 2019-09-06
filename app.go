package main

import (
	"bitbucket.org/scm-manager/plugin-snapshot/center"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const API_JOBS = "https://oss.cloudogu.com/jenkins/job/scm-manager/job/plugins/api/json"

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "configuration file")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatal("use plugin-snapshot directory")
	}

	directory := args[len(args)-1]

	err := createIfNotExists(directory)
	if err != nil {
		log.Fatal("failed to create plugin directory", err)
	}

	config := readConfig(configPath)

	jobs := fetchJobs()
	pluginJobs := filterPluginJobs(jobs.Jobs, config)
	plugins := downloadPlugins(directory, pluginJobs)
	writePluginIndex(directory, plugins)
	err = writePluginJson(directory, plugins)
	if err != nil {
		log.Fatal("failed to create plugin center json", err)
	}
}

func readConfig(path string) Config {
	if path == "" {
		return Config{}
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("failed to read configuration", err)
	}

	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("failed to unmarshal configuration", err)
	}
	return config
}

type Config struct {
	Plugins []string
}

func writePluginIndex(directory string, plugins []Plugin) {
	pluginIndex := make(map[string]Plugin)
	for _, plugin := range plugins {
		pluginIndex[plugin.name] = plugin
	}
	index := make(map[string]interface{})
	index["generated"] = time.Now().String()
	index["plugins"] = pluginIndex

	data, err := yaml.Marshal(index)
	if err != nil {
		log.Fatal("failed to marshal plugin index", err)
	}

	err = ioutil.WriteFile(path.Join(directory, "index.yaml"), data, 0644)
	if err != nil {
		log.Fatal("failed to write plugin index", err)
	}
}

func writePluginJson(directory string, plugins []Plugin) error {
	entries := []center.PluginCenterEntry{}
	for _, plugin := range plugins {
		smp := path.Join(directory, plugin.File)
		descriptor, err := center.ReadDescriptor(smp)
		if err != nil {
			return err
		}
		entry := center.Convert(descriptor)

		data, err := ioutil.ReadFile(smp)
		if err != nil {
			return errors.Wrapf(err, "failed to read smp file: %s", smp)
		}

		entry.Sha256sum = fmt.Sprintf("%x", sha256.Sum256(data))
		entry.Links = center.Links{center.Link{plugin.URL}}

		entries = append(entries, entry)
	}

	center := center.PluginCenter{center.Embedded{entries}}
	data, err := json.MarshalIndent(center, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal plugin center")
	}

	err = ioutil.WriteFile(path.Join(directory, "plugin-center.json"), data, 0644)
	if err != nil {
		log.Fatal("failed to write plugin center json", err)
	}
	return nil
}

func downloadPlugins(directory string, pluginJobs []Job) []Plugin {
	branches := []string{"2.0.0", "2.x", "develop", "master"}

	var plugins []Plugin
	for _, pluginJob := range pluginJobs {
		for _, branch := range branches {

			plugin, err := downloadPlugin(directory, pluginJob, branch)
			if err == nil {
				plugins = append(plugins, plugin)
				break
			}

		}
	}

	return plugins
}

func fetchJobs() Jobs {
	resp, err := http.Get(API_JOBS)
	if err != nil {
		log.Fatal("failed to read jobs from jenkins", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("failed to read body from jobs response", err)
	}

	jobs := Jobs{}
	err = json.Unmarshal(data, &jobs)
	if err != nil {
		log.Fatal("failed to unmarshal jobs response", err)
	}
	return jobs
}

func filterPluginJobs(jobs []Job, config Config) []Job {
	var plugins []Job
	for _, job := range jobs {
		if strings.HasSuffix(job.Name, "plugin") && strings.HasPrefix(job.Name, "scm") {
			if len(config.Plugins) > 0 {
				for _, p := range config.Plugins {
					if job.Name == p {
						plugins = append(plugins, job)
					}
				}
			} else {
				plugins = append(plugins, job)
			}
		}
	}
	return plugins
}

func downloadPlugin(directory string, job Job, branch string) (Plugin, error) {
	plugin := Plugin{}

	build, err := fetchBuild(job, branch)
	if err != nil {
		return plugin, err
	}

	for _, artifact := range build.Artifacts {
		if strings.HasSuffix(artifact.FileName, ".smp") {
			url := downloadArtifact(directory, job, branch, artifact)
			plugin.URL = url
			plugin.File = artifact.FileName
			break
		}
	}

	plugin.name = job.Name
	plugin.Revision = findRevision(job, build)
	plugin.Build = build.Number
	plugin.Branch = branch
	plugin.Repository = "https://bitbucket.org/scm-manager/" + job.Name

	return plugin, nil
}

func fetchBuild(job Job, branch string) (Build, error) {
	jobUrl := job.URL + "/job/" + branch + "/lastSuccessfulBuild/api/json"
	resp, err := http.Get(jobUrl)
	if err != nil {
		log.Fatal("failed to get plugin information", job.Name, err)
	}

	build := Build{}

	if resp.StatusCode != 200 {
		return build, errors.New("branch not found?")
	}

	log.Printf("found %s on branch %s", job.Name, branch)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("failed to read data from build response", err)
	}

	err = json.Unmarshal(data, &build)
	if err != nil {
		log.Fatal("failed to unmarshal build", err)
	}

	return build, nil
}

func findRevision(job Job, build Build) string {
	for _, action := range build.Actions {
		if action.Class == "hudson.plugins.mercurial.MercurialTagAction" {
			return action.MercurialNodeName
		} else if isGitAction(job, action) {
			return action.LastBuiltRevision.SHA1
		}
	}
	return ""
}

func isGitAction(job Job, action Action) bool {
	if action.Class != "hudson.plugins.git.util.BuildData" {
		return false
	}

	for _, remoteUrl := range action.RemoteURLs {
		if strings.Contains(remoteUrl, job.Name) {
			return true
		}
	}

	return false
}

func downloadArtifact(directory string, job Job, branch string, artifact Artifact) string {
	downloadUrl := job.URL + "job/" + branch + "/lastSuccessfulBuild/artifact/" + artifact.RelativePath
	resp, err := http.Get(downloadUrl)
	if err != nil {
		log.Fatal("failed to download artifact", err)
	}

	defer resp.Body.Close()

	file, err := os.Create(path.Join(directory, artifact.FileName))
	if err != nil {
		log.Fatal("failed to create download target", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal("failed to download artifact", err)
	}

	return downloadUrl
}

func createIfNotExists(name string) error {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		log.Printf("create directory %s", name)
		err := os.Mkdir(name, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

type Plugin struct {
	name       string
	Revision   string
	Build      int
	Branch     string
	Repository string
	File       string
	URL        string
}

type Jobs struct {
	Jobs []Job
}

type Job struct {
	Name string
	URL  string
}

type LastBuiltRevision struct {
	SHA1 string
}

type Action struct {
	Class             string `json:"_class"`
	MercurialNodeName string
	LastBuiltRevision LastBuiltRevision
	RemoteURLs        []string
}

type Artifact struct {
	DisplayName  string
	FileName     string
	RelativePath string
}

type Build struct {
	Actions   []Action
	Artifacts []Artifact
	Number    int
}
