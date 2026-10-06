package cfop_view

import (
	entity_public "armazenda/entity/public"
	"armazenda/model/cfop_model"
	"armazenda/model/nfe_model"
	"armazenda/pkg/nfe/defaults"
)

// DefaultCFOP is the system fallback used when the farm has no NFe
// configuration or no default CFOP configured.
const DefaultCFOP = "5101"

// CfopOptions is the selector-ready option set: "Mais utilizados" (farm use
// count > 0, most used first) and "Todos os CFOPs" (code ASC), plus the
// resolved farm default code. The most-used slice is empty when the farm has
// no usage history.
type CfopOptions struct {
	MostUsed []entity_public.CfopOption
	All      []entity_public.CfopOption
	Default  string
}

// BuildCfopOptions groups the merged catalog+farm options and guarantees the
// selected default code is always selectable. When the default is absent from
// the merged list it is prepended to "Todos os CFOPs" with the description
// derived from NaturezaOpForCFOP. A blank default falls back to 5101.
func BuildCfopOptions(options []entity_public.CfopOption, defaultCFOP string) CfopOptions {
	mostUsed, all := cfop_model.GroupCfopOptions(options)

	if defaultCFOP == "" {
		defaultCFOP = DefaultCFOP
	}

	found := false
	for _, option := range all {
		if option.Code == defaultCFOP {
			found = true
			break
		}
	}
	if !found {
		all = append([]entity_public.CfopOption{{
			Code:        defaultCFOP,
			Description: defaults.NaturezaOpForCFOP(defaultCFOP),
		}}, all...)
	}

	return CfopOptions{MostUsed: mostUsed, All: all, Default: defaultCFOP}
}

// GetCfopOptionsWithDefault loads the catalog+farm options for a farm, ranked
// by that farm's use counters, and applies the explicitly resolved default.
func GetCfopOptionsWithDefault(farmID uint32, defaultCFOP string) (CfopOptions, *entity_public.Toast) {
	options, modelErr := cfop_model.GetCfopModel().ListCfopsForFarm(farmID)
	if modelErr != nil {
		toast := entity_public.GetErrorToast("Falhamos ao buscar os CFOPs", "")
		return CfopOptions{}, &toast
	}

	return BuildCfopOptions(options, defaultCFOP), nil
}

// GetCfopOptions resolves the farm's default CFOP (nfe_farm_config.default_cfop,
// else 5101) and loads the ranked option list.
func GetCfopOptions(farmID uint32) (CfopOptions, *entity_public.Toast) {
	defaultCFOP := DefaultCFOP
	if cfg, cfgErr := nfe_model.GetNFeModel().GetFarmConfig(farmID); cfgErr == nil && cfg != nil && cfg.DefaultCFOP != "" {
		defaultCFOP = cfg.DefaultCFOP
	}

	return GetCfopOptionsWithDefault(farmID, defaultCFOP)
}
