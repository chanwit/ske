package addons

import "github.com/rancher/rke/templates"

func GetIstioManifest(Config interface{}) (string, error) {

	return templates.AddonIstioTemplate, nil
}
