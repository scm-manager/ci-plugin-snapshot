package center

type Link struct {
	Href string `json:"name"`
}

type Links struct {
	Download Link `json:"download"`
}

type JsonConditions struct {
	Os         []string `json:"os,omitempty"`
	MinVersion string   `json:"minVersion,omitempty"`
	Arch       string   `json:"arch,omitempty"`
}

type PluginCenterEntry struct {
	Name        string         `json:"name,omitempty"`
	DisplayName string         `json:"displayName,omitempty"`
	Description string         `json:"description,omitempty"`
	Category    string         `json:"category,omitempty"`
	Version     string         `json:"version,omitempty"`
	Author      string         `json:"author,omitempty"`
	Conditions  JsonConditions `json:"conditions,omitempty"`
	Sha256sum   string         `json:"sha256sum,omitempty"`
	Links       Links          `json:"_links"`
}

type Embedded struct {
	Plugins []PluginCenterEntry `json:"plugins"`
}

type PluginCenter struct {
	Embedded Embedded `json:"_embedded"`
}
