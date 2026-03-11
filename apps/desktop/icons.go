package main

import _ "embed"

//go:embed build/appicon.png
var appIconData []byte

//go:embed build/tray-template.png
var trayTemplateData []byte

//go:embed build/tray-template-2x.png
var trayTemplate2xData []byte

//go:embed build/tray-icon.png
var trayIconData []byte
