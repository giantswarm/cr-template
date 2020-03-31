package key

const AppCRTemplate = `
{{ .ConfigmapCR -}}
---
{{ .SecretCR -}}
---
{{ .KubeConfigSecretCR -}}
---
{{ .AppCR -}}
`

const AppCatalogCRTemplate = `
{{ .ConfigmapCR -}}
---
{{ .SecretCR -}}
---
{{ .AppCatalogCR -}}
`

const ClusterCRsTemplate = `
{{ .ClusterCR -}}
---
{{ .AWSClusterCR -}}
{{ if .TemplateDefaultNodepool}}
---
{{ .MachineDeploymentCR -}}
---
{{ .AWSMachineDeploymentCR -}}
{{ end }}
`

const MachineDeploymentCRsTemplate = `
{{ .MachineDeploymentCR -}}
---
{{ .AWSMachineDeploymentCR -}}
`
