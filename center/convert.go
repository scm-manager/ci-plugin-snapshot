package center

func Convert(descriptor PluginDescriptor) PluginCenterEntry {
  return PluginCenterEntry{
    Name:         descriptor.Information.Name,
    Version:      descriptor.Information.Version,
    Category:     descriptor.Information.Category,
    DisplayName:  descriptor.Information.DisplayName,
    Description:  descriptor.Information.Description,
    Author:       descriptor.Information.Author,
    Dependencies: descriptor.Dependencies.Dependency,
    OptionalDependencies: descriptor.OptionalDependencies.OptionalDependency,
    Conditions:   convertConditions(descriptor.Conditions),
  }
}

func convertConditions(conditions Conditions) JsonConditions {
  var os []string
  for _, name := range conditions.Os.Name {
    os = append(os, name)
  }
  return JsonConditions{
    Os:         os,
    MinVersion: conditions.MinVersion,
    Arch:       conditions.Arch,
  }
}
