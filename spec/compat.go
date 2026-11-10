package spec

// Names kept for code written against raff-go v0.8.7 and earlier. The spec's region
// `products` enum (cost_optimized, premium, standard, windows, ...) shares these values, so
// oapi-codegen now prefixes both enums with their type name.

const (
	// Deprecated: use ListTemplatesParamsVMTypeCostOptimized.
	CostOptimized = ListTemplatesParamsVMTypeCostOptimized
	// Deprecated: use ListTemplatesParamsVMTypePremium.
	Premium = ListTemplatesParamsVMTypePremium
	// Deprecated: use ListTemplatesParamsVMTypeStandard.
	Standard = ListTemplatesParamsVMTypeStandard
	// Deprecated: use ListTemplatesParamsVMTypeWindows.
	Windows = ListTemplatesParamsVMTypeWindows
)

// Names kept for code written against raff-go v0.8.8 and earlier. The VM network
// `security_group_source` enum shares the value `manual`, so oapi-codegen now names the
// cluster maintenance upgrade modes without their type prefix.
const (
	// Deprecated: use AutoMinor.
	UpdateK8SClusterMaintenanceJSONBodyUpgradeModeAutoMinor = AutoMinor
	// Deprecated: use AutoPatch.
	UpdateK8SClusterMaintenanceJSONBodyUpgradeModeAutoPatch = AutoPatch
	// Deprecated: use Manual.
	UpdateK8SClusterMaintenanceJSONBodyUpgradeModeManual = Manual
)
